module Fizzy
  class Client
    attr_reader :api_url, :account

    def initialize(token:, api_url:, account: nil)
      @token = token
      @api_url = api_url
      @account = account
    end

    def get(path, params = {})
      uri = build_uri(path, params)
      request = Net::HTTP::Get.new(uri)
      execute(uri, request)
    end

    def post(path, body = {})
      uri = build_uri(path)
      request = Net::HTTP::Post.new(uri)
      request.body = body.to_json unless body.empty?
      request.content_type = "application/json"
      execute(uri, request)
    end

    def put(path, body = {})
      uri = build_uri(path)
      request = Net::HTTP::Put.new(uri)
      request.body = body.to_json unless body.empty?
      request.content_type = "application/json"
      execute(uri, request)
    end

    def delete(path)
      uri = build_uri(path)
      request = Net::HTTP::Delete.new(uri)
      execute(uri, request)
    end

    def account_path(path)
      raise ConfigError, "No account configured. Set FIZZY_ACCOUNT or use --account" unless @account
      "/#{@account}#{path}"
    end

    private

    def build_uri(path, params = {})
      uri = URI.join(@api_url, path)
      uri.query = URI.encode_www_form(params) unless params.empty?
      uri
    end

    def execute(uri, request)
      request["Authorization"] = "Bearer #{@token}"
      request["Accept"] = "application/json"

      response = Net::HTTP.start(uri.host, uri.port, use_ssl: uri.scheme == "https") do |http|
        http.open_timeout = 10
        http.read_timeout = 30
        http.request(request)
      end

      handle_response(response)
    rescue Errno::ECONNREFUSED, Errno::ETIMEDOUT, Net::OpenTimeout, Net::ReadTimeout, SocketError => e
      raise NetworkError, "Connection failed: #{e.message}"
    end

    def handle_response(response)
      case response
      when Net::HTTPSuccess, Net::HTTPCreated, Net::HTTPNoContent
        parse_success_response(response)
      when Net::HTTPUnauthorized
        raise AuthError, "Invalid or expired token"
      when Net::HTTPForbidden
        raise ForbiddenError, parse_error_message(response)
      when Net::HTTPNotFound
        raise NotFoundError, parse_error_message(response)
      when Net::HTTPUnprocessableEntity
        raise ValidationError, parse_error_message(response)
      else
        raise Error, "Request failed: #{response.code} #{response.message}"
      end
    end

    def parse_success_response(response)
      return nil if response.body.nil? || response.body.empty?

      data = JSON.parse(response.body)
      pagination = parse_link_header(response["Link"])

      { data: data, pagination: pagination }
    rescue JSON::ParserError
      { data: response.body, pagination: nil }
    end

    def parse_error_message(response)
      return response.message if response.body.nil? || response.body.empty?

      data = JSON.parse(response.body)
      if data.is_a?(Hash)
        data.map { |k, v| "#{k}: #{Array(v).join(', ')}" }.join("; ")
      else
        data.to_s
      end
    rescue JSON::ParserError
      response.message
    end

    def parse_link_header(header)
      return nil unless header

      links = {}
      header.split(",").each do |link|
        if link =~ /<([^>]+)>;\s*rel="([^"]+)"/
          links[$2] = $1
        end
      end

      return nil if links.empty?

      {
        has_next: links.key?("next"),
        next_url: links["next"]
      }
    end
  end
end

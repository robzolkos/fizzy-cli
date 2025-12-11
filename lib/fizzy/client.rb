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

    def post_multipart(path, params = {}, files = {})
      uri = build_uri(path)
      request = Net::HTTP::Post.new(uri)
      set_multipart_body(request, params, files)
      execute(uri, request)
    end

    def put_multipart(path, params = {}, files = {})
      uri = build_uri(path)
      request = Net::HTTP::Put.new(uri)
      set_multipart_body(request, params, files)
      execute(uri, request)
    end

    def delete(path)
      uri = build_uri(path)
      request = Net::HTTP::Delete.new(uri)
      execute(uri, request)
    end

    # Direct upload for rich text attachments (ActionText)
    # Returns the signed_id to use in <action-text-attachment sgid="...">
    def direct_upload(file_path)
      raise Fizzy::ValidationError, "File not found: #{file_path}" unless File.exist?(file_path)

      file_content = File.binread(file_path)
      filename = File.basename(file_path)
      content_type = detect_content_type(file_path)
      byte_size = file_content.bytesize
      checksum = Base64.strict_encode64(Digest::MD5.digest(file_content))

      # Step 1: Create direct upload
      blob_params = {
        blob: {
          filename: filename,
          byte_size: byte_size,
          checksum: checksum,
          content_type: content_type
        }
      }

      upload_info = post(account_path("/rails/active_storage/direct_uploads"), blob_params)
      raise Fizzy::Error, "Failed to create direct upload" unless upload_info && upload_info[:data]

      data = upload_info[:data]
      direct_upload = data["direct_upload"]
      raise Fizzy::Error, "No direct upload URL returned" unless direct_upload

      # Step 2: Upload file to storage
      upload_uri = URI.parse(direct_upload["url"])
      upload_request = Net::HTTP::Put.new(upload_uri)
      upload_request.body = file_content

      direct_upload["headers"]&.each do |key, value|
        upload_request[key] = value
      end

      upload_response = Net::HTTP.start(upload_uri.host, upload_uri.port, use_ssl: upload_uri.scheme == "https") do |http|
        http.open_timeout = 30
        http.read_timeout = 120
        http.request(upload_request)
      end

      unless upload_response.is_a?(Net::HTTPSuccess)
        raise Fizzy::Error, "Failed to upload file: #{upload_response.code} #{upload_response.message}"
      end

      # Return the signed_id for use in action-text-attachment
      {
        data: {
          signed_id: data["signed_id"],
          filename: data["filename"],
          content_type: data["content_type"],
          byte_size: data["byte_size"]
        }
      }
    end

    def account_path(path)
      raise ConfigError, "No account configured. Set FIZZY_ACCOUNT or use --account" unless @account
      "/#{@account}#{path}"
    end

    def get_all(path, params = {})
      all_data = []
      current_params = params.dup

      loop do
        result = get(path, current_params)
        data = result[:data]
        all_data.concat(Array(data))

        pagination = result[:pagination]
        break unless pagination && pagination[:has_next] && pagination[:next_url]

        next_uri = URI.parse(pagination[:next_url])
        next_params = URI.decode_www_form(next_uri.query || "").to_h
        current_params = next_params.transform_keys(&:to_sym)
      end

      { data: all_data, pagination: nil }
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

    def set_multipart_body(request, params, files)
      boundary = "----FizzyCLI#{SecureRandom.hex(16)}"
      request.content_type = "multipart/form-data; boundary=#{boundary}"

      body = []

      # Add regular params
      params.each do |key, value|
        if value.is_a?(Hash)
          value.each do |nested_key, nested_value|
            body << "--#{boundary}\r\n"
            body << "Content-Disposition: form-data; name=\"#{key}[#{nested_key}]\"\r\n\r\n"
            body << "#{nested_value}\r\n"
          end
        elsif value.is_a?(Array)
          value.each do |item|
            body << "--#{boundary}\r\n"
            body << "Content-Disposition: form-data; name=\"#{key}[]\"\r\n\r\n"
            body << "#{item}\r\n"
          end
        else
          body << "--#{boundary}\r\n"
          body << "Content-Disposition: form-data; name=\"#{key}\"\r\n\r\n"
          body << "#{value}\r\n"
        end
      end

      # Add files
      files.each do |key, file_path|
        next unless file_path && File.exist?(file_path)

        filename = File.basename(file_path)
        content_type = detect_content_type(file_path)
        file_content = File.binread(file_path)

        body << "--#{boundary}\r\n"
        body << "Content-Disposition: form-data; name=\"#{key}\"; filename=\"#{filename}\"\r\n"
        body << "Content-Type: #{content_type}\r\n\r\n"
        body << file_content
        body << "\r\n"
      end

      body << "--#{boundary}--\r\n"
      request.body = body.join
    end

    def detect_content_type(file_path)
      extension = File.extname(file_path).downcase
      case extension
      when ".jpg", ".jpeg"
        "image/jpeg"
      when ".png"
        "image/png"
      when ".gif"
        "image/gif"
      when ".webp"
        "image/webp"
      when ".pdf"
        "application/pdf"
      when ".txt"
        "text/plain"
      else
        "application/octet-stream"
      end
    end
  end
end

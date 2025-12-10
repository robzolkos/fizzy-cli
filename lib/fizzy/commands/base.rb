module Fizzy
  module Commands
    class Base < Thor
      def self.exit_on_failure?
        true
      end

      protected

      def config
        @config ||= Config.new(
          token: parent_options[:token],
          account: parent_options[:account],
          api_url: parent_options[:api_url]
        )
      end

      def client
        raise AuthError, "No token configured. Run 'fizzy auth login' first." unless config.valid?

        @client ||= Client.new(
          token: config.token,
          account: config.account,
          api_url: config.api_url
        )
      end

      def parent_options
        @parent_options ||= parent&.options || {}
      end

      def output_format
        parent_options[:format] || "json"
      end

      def quiet?
        parent_options[:quiet] || false
      end

      def verbose?
        parent_options[:verbose] || false
      end

      def output(result)
        response = if result.is_a?(Response)
          result
        else
          Response.success(data: result[:data], pagination: result[:pagination])
        end

        puts response.to_json
      end

      def output_error(error)
        code = case error
        when AuthError then "AUTH_ERROR"
        when ForbiddenError then "FORBIDDEN"
        when NotFoundError then "NOT_FOUND"
        when ValidationError then "VALIDATION_ERROR"
        when NetworkError then "NETWORK_ERROR"
        when ConfigError then "CONFIG_ERROR"
        else "ERROR"
        end

        status = case error
        when AuthError then 401
        when ForbiddenError then 403
        when NotFoundError then 404
        when ValidationError then 422
        else nil
        end

        response = Response.error(code: code, message: error.message, status: status)
        puts response.to_json

        exit_code = error.respond_to?(:exit_code) ? error.exit_code : EXIT_ERROR
        exit(exit_code)
      end
    end
  end
end

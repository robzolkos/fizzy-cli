module Fizzy
  module Commands
    class Auth < Base
      desc "login TOKEN", "Save API token to config file"
      def login(token)
        config = Config.new
        config.save!(token: token)

        response = Response.success(data: {
          message: "Token saved to #{Config.config_path || Config::CONFIG_PATHS.first}",
          config_path: Config.config_path || Config::CONFIG_PATHS.first
        })
        puts response.to_json
      rescue StandardError => e
        output_error(Error.new(e.message))
      end

      desc "logout", "Remove stored credentials"
      def logout
        config = Config.new
        config_path = Config.config_path

        if config_path
          config.clear!
          response = Response.success(data: { message: "Credentials removed" })
        else
          response = Response.success(data: { message: "No credentials found" })
        end

        puts response.to_json
      rescue StandardError => e
        output_error(Error.new(e.message))
      end

      desc "status", "Show current authentication state"
      def status
        config = Config.new

        data = {
          authenticated: config.valid?,
          config_path: Config.config_path,
          account: config.account,
          api_url: config.api_url
        }

        if config.valid?
          data[:token_preview] = "#{config.token[0..7]}..."
        end

        response = Response.success(data: data)
        puts response.to_json
      end
    end
  end
end

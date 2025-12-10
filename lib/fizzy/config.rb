module Fizzy
  class Config
    DEFAULT_API_URL = "https://app.fizzy.do"

    def self.config_paths
      [
        File.expand_path("~/.fizzy/config.yaml"),
        File.expand_path("~/.config/fizzy/config.yaml")
      ]
    end

    attr_reader :token, :account, :api_url

    def initialize(token: nil, account: nil, api_url: nil)
      file_config = load_config_file

      @token = token || ENV["FIZZY_TOKEN"] || file_config["token"]
      @account = account || ENV["FIZZY_ACCOUNT"] || file_config["account"]
      @api_url = api_url || ENV["FIZZY_API_URL"] || file_config["api_url"] || DEFAULT_API_URL
    end

    def valid?
      !@token.nil? && !@token.empty?
    end

    def save!(token:, account: nil, api_url: nil)
      config_path = self.class.config_paths.first
      config_dir = File.dirname(config_path)

      FileUtils.mkdir_p(config_dir) unless Dir.exist?(config_dir)

      config = {}
      config["token"] = token
      config["account"] = account if account
      config["api_url"] = api_url if api_url && api_url != DEFAULT_API_URL

      File.write(config_path, config.to_yaml)
      File.chmod(0600, config_path)

      @token = token
      @account = account if account
      @api_url = api_url if api_url
    end

    def clear!
      self.class.config_paths.each do |path|
        File.delete(path) if File.exist?(path)
      end
      @token = nil
      @account = nil
      @api_url = DEFAULT_API_URL
    end

    def self.config_path
      config_paths.find { |path| File.exist?(path) }
    end

    private

    def load_config_file
      config_path = self.class.config_path
      return {} unless config_path

      YAML.load_file(config_path) || {}
    rescue Errno::ENOENT, Psych::SyntaxError
      {}
    end
  end
end

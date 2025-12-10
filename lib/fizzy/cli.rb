module Fizzy
  class CLI < Thor
    def self.exit_on_failure?
      true
    end

    class_option :token, type: :string, desc: "API access token (or FIZZY_TOKEN env var)"
    class_option :account, type: :string, desc: "Account ID (or FIZZY_ACCOUNT env var)"
    class_option :api_url, type: :string, desc: "API base URL (default: https://app.fizzy.do)"
    class_option :format, type: :string, default: "json", desc: "Output format: json, text"
    class_option :quiet, type: :boolean, default: false, desc: "Suppress non-essential output"
    class_option :verbose, type: :boolean, default: false, desc: "Show request/response details"

    desc "auth SUBCOMMAND", "Manage authentication"
    subcommand "auth", Commands::Auth

    desc "identity SUBCOMMAND", "Manage identity"
    subcommand "identity", Commands::Identity
  end
end

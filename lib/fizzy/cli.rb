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

    desc "board SUBCOMMAND", "Manage boards"
    subcommand "board", Commands::Board

    desc "card SUBCOMMAND", "Manage cards"
    subcommand "card", Commands::Card

    desc "column SUBCOMMAND", "Manage columns"
    subcommand "column", Commands::Column

    desc "user SUBCOMMAND", "Manage users"
    subcommand "user", Commands::User

    desc "tag SUBCOMMAND", "Manage tags"
    subcommand "tag", Commands::Tag

    desc "comment SUBCOMMAND", "Manage comments"
    subcommand "comment", Commands::Comment

    desc "reaction SUBCOMMAND", "Manage reactions"
    subcommand "reaction", Commands::Reaction

    desc "step SUBCOMMAND", "Manage steps (to-do items)"
    subcommand "step", Commands::Step

    desc "notification SUBCOMMAND", "Manage notifications"
    subcommand "notification", Commands::Notification

    desc "upload SUBCOMMAND", "Upload files for rich text"
    subcommand "upload", Commands::Upload
  end
end

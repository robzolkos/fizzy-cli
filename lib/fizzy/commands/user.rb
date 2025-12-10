module Fizzy
  module Commands
    class User < Base
      desc "list", "List all users"
      option :page, type: :numeric, desc: "Page number"
      option :all, type: :boolean, default: false, desc: "Fetch all pages"
      def list
        params = {}
        params[:page] = options[:page] if options[:page]

        result = if options[:all]
          client.get_all(client.account_path("/users"), params)
        else
          client.get(client.account_path("/users"), params)
        end
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "show ID", "Show a specific user"
      def show(id)
        result = client.get(client.account_path("/users/#{id}"))
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "update ID", "Update a user"
      option :name, type: :string, desc: "User name"
      option :avatar, type: :string, desc: "Path to avatar image file"
      def update(id)
        user_params = {}
        user_params[:name] = options[:name] if options.key?(:name)

        result = if options[:avatar]
          client.put_multipart(
            client.account_path("/users/#{id}"),
            { user: user_params },
            { "user[avatar]" => options[:avatar] }
          )
        else
          client.put(client.account_path("/users/#{id}"), { user: user_params })
        end
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "deactivate ID", "Deactivate a user"
      def deactivate(id)
        result = client.delete(client.account_path("/users/#{id}"))
        output(result || Response.success(data: { deactivated: true }))
      rescue Fizzy::Error => e
        output_error(e)
      end
    end
  end
end

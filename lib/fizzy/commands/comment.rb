module Fizzy
  module Commands
    class Comment < Base
      desc "list", "List comments on a card"
      option :card, required: true, type: :string, desc: "Card number"
      option :page, type: :numeric, desc: "Page number"
      option :all, type: :boolean, default: false, desc: "Fetch all pages"
      def list
        params = {}
        params[:page] = options[:page] if options[:page]

        result = if options[:all]
          client.get_all(client.account_path("/cards/#{options[:card]}/comments"), params)
        else
          client.get(client.account_path("/cards/#{options[:card]}/comments"), params)
        end
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "show ID", "Show a specific comment"
      option :card, required: true, type: :string, desc: "Card number"
      def show(id)
        result = client.get(client.account_path("/cards/#{options[:card]}/comments/#{id}"))
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "create", "Create a new comment"
      option :card, required: true, type: :string, desc: "Card number"
      option :body, type: :string, desc: "Comment body (supports rich text)"
      option :body_file, type: :string, desc: "Read body from file"
      option :created_at, type: :string, desc: "Custom creation timestamp (ISO 8601)"
      def create
        comment_params = {}

        if options[:body_file]
          comment_params[:body] = File.read(options[:body_file])
        elsif options[:body]
          comment_params[:body] = options[:body]
        else
          raise Fizzy::ValidationError, "Either --body or --body-file is required"
        end

        comment_params[:created_at] = options[:created_at] if options[:created_at]

        result = client.post(client.account_path("/cards/#{options[:card]}/comments"), { comment: comment_params })
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "update ID", "Update a comment"
      option :card, required: true, type: :string, desc: "Card number"
      option :body, type: :string, desc: "Comment body (supports rich text)"
      option :body_file, type: :string, desc: "Read body from file"
      option :created_at, type: :string, desc: "Custom creation timestamp (ISO 8601)"
      def update(id)
        comment_params = {}

        if options[:body_file]
          comment_params[:body] = File.read(options[:body_file])
        elsif options[:body]
          comment_params[:body] = options[:body]
        end

        comment_params[:created_at] = options[:created_at] if options[:created_at]

        result = client.put(client.account_path("/cards/#{options[:card]}/comments/#{id}"), { comment: comment_params })
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "delete ID", "Delete a comment"
      option :card, required: true, type: :string, desc: "Card number"
      def delete(id)
        result = client.delete(client.account_path("/cards/#{options[:card]}/comments/#{id}"))
        output(result || Response.success(data: { deleted: true }))
      rescue Fizzy::Error => e
        output_error(e)
      end
    end
  end
end

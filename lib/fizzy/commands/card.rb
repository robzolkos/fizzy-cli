module Fizzy
  module Commands
    class Card < Base
      desc "list", "List cards"
      option :board, type: :string, desc: "Filter by board ID"
      option :tag, type: :string, desc: "Filter by tag ID"
      option :assignee, type: :string, desc: "Filter by assignee user ID"
      option :status, type: :string, desc: "Filter by status (published, closed, not_now)"
      option :page, type: :numeric, desc: "Page number"
      option :all, type: :boolean, default: false, desc: "Fetch all pages"
      def list
        params = {}
        params["board_ids[]"] = options[:board] if options[:board]
        params["tag_ids[]"] = options[:tag] if options[:tag]
        params["assignee_ids[]"] = options[:assignee] if options[:assignee]
        params[:status] = options[:status] if options[:status]
        params[:page] = options[:page] if options[:page]

        result = if options[:all]
          client.get_all(client.account_path("/cards"), params)
        else
          client.get(client.account_path("/cards"), params)
        end
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "show NUMBER", "Show a specific card by number"
      def show(number)
        result = client.get(client.account_path("/cards/#{number}"))
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "create", "Create a new card"
      option :board, required: true, type: :string, desc: "Board ID"
      option :title, required: true, type: :string, desc: "Card title"
      option :description, type: :string, desc: "Card description (rich text/HTML)"
      option :description_file, type: :string, desc: "Read description from file"
      option :status, type: :string, desc: "Card status"
      option :tag_ids, type: :string, desc: "Comma-separated tag IDs"
      option :image, type: :string, desc: "Path to header image file"
      def create
        card_params = {
          title: options[:title]
        }

        if options[:description_file]
          card_params[:description] = File.read(options[:description_file])
        elsif options[:description]
          card_params[:description] = options[:description]
        end

        card_params[:status] = options[:status] if options[:status]

        if options[:tag_ids]
          card_params[:tag_ids] = options[:tag_ids].split(",").map(&:strip)
        end

        result = if options[:image]
          client.post_multipart(
            client.account_path("/boards/#{options[:board]}/cards"),
            { card: card_params },
            { "card[image]" => options[:image] }
          )
        else
          client.post(client.account_path("/boards/#{options[:board]}/cards"), { card: card_params })
        end
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "update NUMBER", "Update a card"
      option :title, type: :string, desc: "Card title"
      option :description, type: :string, desc: "Card description (rich text/HTML)"
      option :description_file, type: :string, desc: "Read description from file"
      option :status, type: :string, desc: "Card status"
      option :tag_ids, type: :string, desc: "Comma-separated tag IDs"
      option :image, type: :string, desc: "Path to header image file"
      def update(number)
        card_params = {}
        card_params[:title] = options[:title] if options.key?(:title)
        card_params[:status] = options[:status] if options.key?(:status)

        if options[:description_file]
          card_params[:description] = File.read(options[:description_file])
        elsif options.key?(:description)
          card_params[:description] = options[:description]
        end

        if options[:tag_ids]
          card_params[:tag_ids] = options[:tag_ids].split(",").map(&:strip)
        end

        result = if options[:image]
          client.put_multipart(
            client.account_path("/cards/#{number}"),
            { card: card_params },
            { "card[image]" => options[:image] }
          )
        else
          client.put(client.account_path("/cards/#{number}"), { card: card_params })
        end
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "delete NUMBER", "Delete a card"
      def delete(number)
        result = client.delete(client.account_path("/cards/#{number}"))
        output(result || Response.success(data: { deleted: true }))
      rescue Fizzy::Error => e
        output_error(e)
      end

      # Card Action Commands

      desc "close NUMBER", "Close a card"
      def close(number)
        result = client.post(client.account_path("/cards/#{number}/closure"), {})
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "reopen NUMBER", "Reopen a closed card"
      def reopen(number)
        result = client.delete(client.account_path("/cards/#{number}/closure"))
        output(result || Response.success(data: { reopened: true }))
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "postpone NUMBER", "Postpone a card (mark as not now)"
      def postpone(number)
        result = client.post(client.account_path("/cards/#{number}/not_now"), {})
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "column NUMBER", "Move a card into a column"
      option :column, required: true, type: :string, desc: "Column ID to move into"
      def column(number)
        result = client.post(client.account_path("/cards/#{number}/triage"), { column_id: options[:column] })
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "untriage NUMBER", "Send a card back to triage"
      def untriage(number)
        result = client.delete(client.account_path("/cards/#{number}/triage"))
        output(result || Response.success(data: { untriaged: true }))
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "assign NUMBER", "Toggle assignment for a user on a card"
      option :user, required: true, type: :string, desc: "User ID to toggle assignment"
      def assign(number)
        result = client.post(client.account_path("/cards/#{number}/assignments"), { assignee_id: options[:user] })
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "tag NUMBER", "Toggle a tag on a card (creates tag if needed)"
      option :tag, required: true, type: :string, desc: "Tag title to toggle"
      def tag(number)
        result = client.post(client.account_path("/cards/#{number}/taggings"), { tag_title: options[:tag] })
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "watch NUMBER", "Watch a card for notifications"
      def watch(number)
        result = client.post(client.account_path("/cards/#{number}/watch"), {})
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end

      desc "unwatch NUMBER", "Stop watching a card"
      def unwatch(number)
        result = client.delete(client.account_path("/cards/#{number}/watch"))
        output(result || Response.success(data: { unwatched: true }))
      rescue Fizzy::Error => e
        output_error(e)
      end
    end
  end
end

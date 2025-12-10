module Fizzy
  module Commands
    class Identity < Base
      desc "show", "Show your identity and accessible accounts"
      def show
        result = client.get("/my/identity")
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end
    end
  end
end

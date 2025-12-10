module Fizzy
  module Commands
    class Upload < Base
      desc "file PATH", "Upload a file for use in rich text fields"
      long_desc <<-DESC
        Uploads a file using Active Storage direct upload and returns the signed_id.

        The signed_id can be used in rich text fields (card descriptions, comment bodies)
        by embedding it in an action-text-attachment tag:

        <action-text-attachment sgid="SIGNED_ID"></action-text-attachment>

        Example workflow:

        1. Upload the file:
           $ fizzy upload file /path/to/image.png
           # Returns: {"signed_id": "eyJfcmFpbHMi..."}

        2. Use the signed_id in a card description:
           $ fizzy card create --board BOARD_ID --title "My Card" \\
               --description '<p>See image:</p><action-text-attachment sgid="eyJfcmFpbHMi..."></action-text-attachment>'
      DESC
      def file(path)
        raise Fizzy::ValidationError, "File not found: #{path}" unless File.exist?(path)

        result = client.direct_upload(path)
        output(result)
      rescue Fizzy::Error => e
        output_error(e)
      end
    end
  end
end

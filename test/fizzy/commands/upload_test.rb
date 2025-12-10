require "test_helper"

class Fizzy::Commands::UploadTest < Fizzy::TestCase
  def setup
    super
    @original_home = ENV["HOME"]
    @temp_dir = Dir.mktmpdir
    ENV["HOME"] = @temp_dir

    config = Fizzy::Config.new
    config.save!(token: "test_token", account: "test_account")

    @test_image = File.join(File.dirname(__FILE__), "..", "..", "fixtures", "files", "test_image.png")
  end

  def teardown
    super
    ENV["HOME"] = @original_home
    FileUtils.rm_rf(@temp_dir)
  end

  def test_upload_file_requires_existing_file
    output = capture_output do
      begin
        Fizzy::Commands::Upload.new.invoke(:file, ["/nonexistent/file.png"])
      rescue SystemExit
        # Expected
      end
    end

    result = JSON.parse(output)
    refute result["success"]
    assert_equal "VALIDATION_ERROR", result["error"]["code"]
    assert_match(/not found/i, result["error"]["message"])
  end

  def test_upload_file_creates_direct_upload
    # Step 1: Create direct upload
    stub_request(:post, "https://app.fizzy.do/rails/active_storage/direct_uploads")
      .to_return(
        status: 200,
        body: {
          id: "blob123",
          key: "abc123",
          filename: "test_image.png",
          content_type: "image/png",
          byte_size: 321,
          checksum: "abc123==",
          signed_id: "eyJfcmFpbHMi...",
          direct_upload: {
            url: "https://storage.example.com/upload",
            headers: {
              "Content-Type" => "image/png",
              "Content-MD5" => "abc123=="
            }
          }
        }.to_json,
        headers: { "Content-Type" => "application/json" }
      )

    # Step 2: Upload to storage
    stub_request(:put, "https://storage.example.com/upload")
      .to_return(status: 200, body: "")

    output = capture_output do
      Fizzy::Commands::Upload.new.invoke(:file, [@test_image])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "eyJfcmFpbHMi...", result["data"]["signed_id"]
    assert_equal "test_image.png", result["data"]["filename"]
  end

  private

  def capture_output
    original_stdout = $stdout
    $stdout = StringIO.new
    yield
    $stdout.string
  ensure
    $stdout = original_stdout
  end
end

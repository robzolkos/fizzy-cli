require "test_helper"

class Fizzy::Commands::IdentityTest < Fizzy::TestCase
  def setup
    super
    @original_home = ENV["HOME"]
    @temp_dir = Dir.mktmpdir
    ENV["HOME"] = @temp_dir

    # Save a token so we're authenticated
    config = Fizzy::Config.new
    config.save!(token: "test_token")
  end

  def teardown
    super
    ENV["HOME"] = @original_home
    FileUtils.rm_rf(@temp_dir)
  end

  def test_show_returns_identity
    stub_request(:get, "https://app.fizzy.do/my/identity")
      .with(headers: { "Authorization" => "Bearer test_token" })
      .to_return(
        status: 200,
        body: '{"accounts": [{"id": "123", "name": "Test Account"}]}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Identity.new.invoke(:show, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal 1, result["data"]["accounts"].length
    assert_equal "Test Account", result["data"]["accounts"][0]["name"]
  end

  def test_show_with_invalid_token
    stub_request(:get, "https://app.fizzy.do/my/identity")
      .to_return(status: 401, body: '{"error": "Unauthorized"}')

    assert_raises(SystemExit) do
      capture_output do
        Fizzy::Commands::Identity.new.invoke(:show, [])
      end
    end
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

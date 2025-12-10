require "test_helper"

class Fizzy::Commands::AuthTest < Fizzy::TestCase
  def setup
    super
    @original_home = ENV["HOME"]
    @temp_dir = Dir.mktmpdir
    ENV["HOME"] = @temp_dir
  end

  def teardown
    super
    ENV["HOME"] = @original_home
    FileUtils.rm_rf(@temp_dir)
  end

  def test_login_saves_token
    output = capture_output do
      Fizzy::Commands::Auth.new.invoke(:login, ["test_token_123"])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_includes result["data"]["message"], "Token saved"

    # Verify token was saved
    config = Fizzy::Config.new
    assert_equal "test_token_123", config.token
  end

  def test_logout_removes_credentials
    # First save a token
    config = Fizzy::Config.new
    config.save!(token: "token_to_remove")

    output = capture_output do
      Fizzy::Commands::Auth.new.invoke(:logout, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "Credentials removed", result["data"]["message"]

    # Verify token was removed
    new_config = Fizzy::Config.new
    refute new_config.valid?
  end

  def test_logout_when_no_credentials
    output = capture_output do
      Fizzy::Commands::Auth.new.invoke(:logout, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "No credentials found", result["data"]["message"]
  end

  def test_status_when_not_authenticated
    output = capture_output do
      Fizzy::Commands::Auth.new.invoke(:status, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    refute result["data"]["authenticated"]
    assert_nil result["data"]["config_path"]
  end

  def test_status_when_authenticated
    config = Fizzy::Config.new
    config.save!(token: "my_secret_token")

    output = capture_output do
      Fizzy::Commands::Auth.new.invoke(:status, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert result["data"]["authenticated"]
    assert result["data"]["config_path"]
    assert_equal "my_secre...", result["data"]["token_preview"]
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

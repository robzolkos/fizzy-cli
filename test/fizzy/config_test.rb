require "test_helper"

class Fizzy::ConfigTest < Fizzy::TestCase
  def setup
    super
    @original_env = ENV.to_h
    ENV.delete("FIZZY_TOKEN")
    ENV.delete("FIZZY_ACCOUNT")
    ENV.delete("FIZZY_API_URL")
    # Use temp dir to avoid reading real config
    @original_home = ENV["HOME"]
    @temp_dir = Dir.mktmpdir
    ENV["HOME"] = @temp_dir
  end

  def teardown
    super
    ENV.replace(@original_env)
    ENV["HOME"] = @original_home
    FileUtils.rm_rf(@temp_dir)
  end

  def test_default_api_url
    config = Fizzy::Config.new
    assert_equal "https://app.fizzy.do", config.api_url
  end

  def test_token_from_argument
    config = Fizzy::Config.new(token: "test_token")
    assert_equal "test_token", config.token
    assert config.valid?
  end

  def test_token_from_env
    ENV["FIZZY_TOKEN"] = "env_token"
    config = Fizzy::Config.new
    assert_equal "env_token", config.token
    assert config.valid?
  end

  def test_argument_overrides_env
    ENV["FIZZY_TOKEN"] = "env_token"
    config = Fizzy::Config.new(token: "arg_token")
    assert_equal "arg_token", config.token
  end

  def test_account_from_env
    ENV["FIZZY_ACCOUNT"] = "123456"
    config = Fizzy::Config.new
    assert_equal "123456", config.account
  end

  def test_api_url_from_env
    ENV["FIZZY_API_URL"] = "https://custom.api.com"
    config = Fizzy::Config.new
    assert_equal "https://custom.api.com", config.api_url
  end

  def test_invalid_without_token
    config = Fizzy::Config.new
    refute config.valid?
  end

  def test_invalid_with_empty_token
    config = Fizzy::Config.new(token: "")
    refute config.valid?
  end
end

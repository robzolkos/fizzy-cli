$LOAD_PATH.unshift File.expand_path("../lib", __dir__)

require "minitest/autorun"
require "webmock/minitest"
require "tmpdir"
require "fizzy"

class Fizzy::TestCase < Minitest::Test
  def setup
    WebMock.disable_net_connect!
  end

  def teardown
    WebMock.reset!
  end

  def fixture_path(name)
    File.join(__dir__, "fixtures", name)
  end

  def load_fixture(name)
    File.read(fixture_path(name))
  end

  def json_fixture(name)
    JSON.parse(load_fixture("responses/#{name}.json"))
  end
end

require "test_helper"

class Fizzy::ClientTest < Fizzy::TestCase
  def setup
    super
    @client = Fizzy::Client.new(
      token: "test_token",
      api_url: "https://app.fizzy.do",
      account: "123456"
    )
  end

  def test_get_request
    stub_request(:get, "https://app.fizzy.do/my/identity")
      .with(headers: { "Authorization" => "Bearer test_token", "Accept" => "application/json" })
      .to_return(status: 200, body: '{"accounts": []}', headers: { "Content-Type" => "application/json" })

    result = @client.get("/my/identity")
    assert_equal({ "accounts" => [] }, result[:data])
  end

  def test_get_with_params
    stub_request(:get, "https://app.fizzy.do/123456/cards?status=published")
      .to_return(status: 200, body: '[]', headers: { "Content-Type" => "application/json" })

    result = @client.get("/123456/cards", status: "published")
    assert_equal [], result[:data]
  end

  def test_post_request
    stub_request(:post, "https://app.fizzy.do/123456/boards")
      .with(
        body: '{"board":{"name":"Test"}}',
        headers: { "Content-Type" => "application/json" }
      )
      .to_return(status: 201, body: '{"id": "abc123"}', headers: { "Content-Type" => "application/json" })

    result = @client.post("/123456/boards", { board: { name: "Test" } })
    assert_equal({ "id" => "abc123" }, result[:data])
  end

  def test_put_request
    stub_request(:put, "https://app.fizzy.do/123456/boards/abc")
      .to_return(status: 204, body: nil)

    result = @client.put("/123456/boards/abc", { board: { name: "Updated" } })
    assert_nil result
  end

  def test_delete_request
    stub_request(:delete, "https://app.fizzy.do/123456/boards/abc")
      .to_return(status: 204, body: nil)

    result = @client.delete("/123456/boards/abc")
    assert_nil result
  end

  def test_unauthorized_raises_auth_error
    stub_request(:get, "https://app.fizzy.do/my/identity")
      .to_return(status: 401, body: '{"error": "Unauthorized"}')

    assert_raises(Fizzy::AuthError) do
      @client.get("/my/identity")
    end
  end

  def test_forbidden_raises_forbidden_error
    stub_request(:get, "https://app.fizzy.do/123456/boards/secret")
      .to_return(status: 403, body: '{"error": "Forbidden"}')

    assert_raises(Fizzy::ForbiddenError) do
      @client.get("/123456/boards/secret")
    end
  end

  def test_not_found_raises_not_found_error
    stub_request(:get, "https://app.fizzy.do/123456/cards/999")
      .to_return(status: 404, body: '{"error": "Not found"}')

    assert_raises(Fizzy::NotFoundError) do
      @client.get("/123456/cards/999")
    end
  end

  def test_validation_error_raises_validation_error
    stub_request(:post, "https://app.fizzy.do/123456/boards")
      .to_return(status: 422, body: '{"name": ["can\'t be blank"]}')

    error = assert_raises(Fizzy::ValidationError) do
      @client.post("/123456/boards", { board: {} })
    end
    assert_includes error.message, "name"
  end

  def test_account_path
    assert_equal "/123456/boards", @client.account_path("/boards")
  end

  def test_account_path_without_account_raises_error
    client = Fizzy::Client.new(token: "test", api_url: "https://app.fizzy.do")
    assert_raises(Fizzy::ConfigError) do
      client.account_path("/boards")
    end
  end

  def test_parses_link_header_for_pagination
    stub_request(:get, "https://app.fizzy.do/123456/cards")
      .to_return(
        status: 200,
        body: '[]',
        headers: {
          "Content-Type" => "application/json",
          "Link" => '<https://app.fizzy.do/123456/cards?page=2>; rel="next"'
        }
      )

    result = @client.get("/123456/cards")
    assert result[:pagination][:has_next]
    assert_equal "https://app.fizzy.do/123456/cards?page=2", result[:pagination][:next_url]
  end

  def test_no_pagination_without_link_header
    stub_request(:get, "https://app.fizzy.do/123456/cards")
      .to_return(status: 200, body: '[]', headers: { "Content-Type" => "application/json" })

    result = @client.get("/123456/cards")
    assert_nil result[:pagination]
  end
end

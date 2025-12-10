require "test_helper"

class Fizzy::ResponseTest < Fizzy::TestCase
  def test_success_response
    response = Fizzy::Response.success(data: { id: "123" })

    assert response.success
    assert_equal({ id: "123" }, response.data)
    assert_nil response.error
    assert response.meta[:timestamp]
  end

  def test_success_response_with_pagination
    response = Fizzy::Response.success(
      data: [{ id: "1" }],
      pagination: { has_next: true, next_url: "http://example.com?page=2" }
    )

    assert response.success
    assert response.pagination[:has_next]
  end

  def test_error_response
    response = Fizzy::Response.error(
      code: "NOT_FOUND",
      message: "Card not found",
      status: 404
    )

    refute response.success
    assert_nil response.data
    assert_equal "NOT_FOUND", response.error[:code]
    assert_equal "Card not found", response.error[:message]
    assert_equal 404, response.error[:status]
  end

  def test_to_h
    response = Fizzy::Response.success(data: { name: "Test" })
    hash = response.to_h

    assert hash[:success]
    assert_equal({ name: "Test" }, hash[:data])
    assert hash[:meta][:timestamp]
    refute hash.key?(:error)
    refute hash.key?(:pagination)
  end

  def test_to_json
    response = Fizzy::Response.success(data: { name: "Test" })
    json = response.to_json
    parsed = JSON.parse(json)

    assert parsed["success"]
    assert_equal({ "name" => "Test" }, parsed["data"])
  end

  def test_error_response_to_json
    response = Fizzy::Response.error(code: "ERROR", message: "Something went wrong")
    json = response.to_json
    parsed = JSON.parse(json)

    refute parsed["success"]
    assert_equal "ERROR", parsed["error"]["code"]
    assert_equal "Something went wrong", parsed["error"]["message"]
  end
end

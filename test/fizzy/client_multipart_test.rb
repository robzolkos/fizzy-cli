require "test_helper"

class Fizzy::ClientMultipartTest < Fizzy::TestCase
  def setup
    super
    @client = Fizzy::Client.new(
      token: "test_token",
      api_url: "https://app.fizzy.do",
      account: "test_account"
    )
    @test_image = File.join(__dir__, "..", "fixtures", "files", "test_image.png")
  end

  def test_post_multipart_sends_file
    stub_request(:post, "https://app.fizzy.do/test_account/boards/5/cards")
      .with { |request|
        request.headers["Content-Type"].include?("multipart/form-data") &&
          request.body.include?("test_image.png") &&
          request.body.include?("card[title]")
      }
      .to_return(
        status: 201,
        body: '{"id": "1", "title": "Test"}',
        headers: { "Content-Type" => "application/json" }
      )

    result = @client.post_multipart(
      @client.account_path("/boards/5/cards"),
      { card: { title: "Test Card" } },
      { "card[image]" => @test_image }
    )

    assert result[:data]["id"]
  end

  def test_put_multipart_sends_file
    stub_request(:put, "https://app.fizzy.do/test_account/users/123")
      .with { |request|
        request.headers["Content-Type"].include?("multipart/form-data") &&
          request.body.include?("test_image.png")
      }
      .to_return(
        status: 200,
        body: '{"id": "123", "name": "Test User"}',
        headers: { "Content-Type" => "application/json" }
      )

    result = @client.put_multipart(
      @client.account_path("/users/123"),
      { user: { name: "Test User" } },
      { "user[avatar]" => @test_image }
    )

    assert result[:data]["id"]
  end

  def test_multipart_skips_missing_files
    stub_request(:post, "https://app.fizzy.do/test_account/cards")
      .with { |request|
        request.headers["Content-Type"].include?("multipart/form-data") &&
          !request.body.include?("nonexistent.png")
      }
      .to_return(
        status: 201,
        body: '{"id": "1"}',
        headers: { "Content-Type" => "application/json" }
      )

    result = @client.post_multipart(
      @client.account_path("/cards"),
      { card: { title: "Test" } },
      { "card[image]" => "/nonexistent/path/file.png" }
    )

    assert result[:data]["id"]
  end

  def test_detect_content_type_for_images
    client = Fizzy::Client.new(token: "t", api_url: "https://test.com")

    assert_equal "image/png", client.send(:detect_content_type, "file.png")
    assert_equal "image/jpeg", client.send(:detect_content_type, "file.jpg")
    assert_equal "image/jpeg", client.send(:detect_content_type, "file.jpeg")
    assert_equal "image/gif", client.send(:detect_content_type, "file.gif")
    assert_equal "image/webp", client.send(:detect_content_type, "file.webp")
  end

  def test_detect_content_type_for_documents
    client = Fizzy::Client.new(token: "t", api_url: "https://test.com")

    assert_equal "application/pdf", client.send(:detect_content_type, "doc.pdf")
    assert_equal "text/plain", client.send(:detect_content_type, "file.txt")
    assert_equal "application/octet-stream", client.send(:detect_content_type, "file.unknown")
  end
end

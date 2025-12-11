require "test_helper"

class Fizzy::Commands::CommentTest < Fizzy::TestCase
  def setup
    super
    @original_home = ENV["HOME"]
    @temp_dir = Dir.mktmpdir
    ENV["HOME"] = @temp_dir

    config = Fizzy::Config.new
    config.save!(token: "test_token", account: "test_account")
  end

  def teardown
    super
    ENV["HOME"] = @original_home
    FileUtils.rm_rf(@temp_dir)
  end

  def test_list_returns_comments
    stub_request(:get, "https://app.fizzy.do/test_account/cards/42/comments")
      .to_return(
        status: 200,
        body: '[{"id": "c1", "body": {"plain_text": "First comment"}}]',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42" }).invoke(:list, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal 1, result["data"].length
    assert_equal "First comment", result["data"][0]["body"]["plain_text"]
  end

  def test_list_with_pagination
    stub_request(:get, "https://app.fizzy.do/test_account/cards/42/comments")
      .with(query: { "page" => "2" })
      .to_return(
        status: 200,
        body: '[{"id": "c2", "body": {"plain_text": "Page 2 comment"}}]',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42", page: 2 }).invoke(:list, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "Page 2 comment", result["data"][0]["body"]["plain_text"]
  end

  def test_show_returns_comment
    stub_request(:get, "https://app.fizzy.do/test_account/cards/42/comments/c1")
      .to_return(
        status: 200,
        body: '{"id": "c1", "body": {"plain_text": "Comment details"}}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42" }).invoke(:show, ["c1"])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "c1", result["data"]["id"]
  end

  def test_create_comment
    stub_request(:post, "https://app.fizzy.do/test_account/cards/42/comments")
      .with(
        body: { comment: { body: "New comment" } }.to_json,
        headers: { "Content-Type" => "application/json" }
      )
      .to_return(
        status: 201,
        body: '{"id": "c2", "body": {"plain_text": "New comment"}}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42", body: "New comment" }).invoke(:create, [])
    end

    result = JSON.parse(output)
    assert result["success"]
  end

  def test_create_comment_from_file
    file_path = File.join(@temp_dir, "comment.txt")
    File.write(file_path, "Comment from file")

    stub_request(:post, "https://app.fizzy.do/test_account/cards/42/comments")
      .with(
        body: { comment: { body: "Comment from file" } }.to_json
      )
      .to_return(
        status: 201,
        body: '{"id": "c3"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42", body_file: file_path }).invoke(:create, [])
    end

    result = JSON.parse(output)
    assert result["success"]
  end

  def test_create_comment_requires_body
    output = capture_output do
      begin
        Fizzy::Commands::Comment.new([], { card: "42" }).invoke(:create, [])
      rescue SystemExit
        # Expected
      end
    end

    result = JSON.parse(output)
    refute result["success"]
    assert_equal "VALIDATION_ERROR", result["error"]["code"]
  end

  def test_create_comment_with_created_at
    stub_request(:post, "https://app.fizzy.do/test_account/cards/42/comments")
      .with(
        body: { comment: { body: "New comment", created_at: "2024-01-15T10:30:00Z" } }.to_json,
        headers: { "Content-Type" => "application/json" }
      )
      .to_return(
        status: 201,
        body: '{"id": "c2", "body": {"plain_text": "New comment"}, "created_at": "2024-01-15T10:30:00Z"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42", body: "New comment", created_at: "2024-01-15T10:30:00Z" }).invoke(:create, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "2024-01-15T10:30:00Z", result["data"]["created_at"]
  end

  def test_update_comment
    stub_request(:put, "https://app.fizzy.do/test_account/cards/42/comments/c1")
      .with(
        body: { comment: { body: "Updated body" } }.to_json
      )
      .to_return(
        status: 200,
        body: '{"id": "c1", "body": {"plain_text": "Updated body"}}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42", body: "Updated body" }).invoke(:update, ["c1"])
    end

    result = JSON.parse(output)
    assert result["success"]
  end

  def test_update_comment_with_created_at
    stub_request(:put, "https://app.fizzy.do/test_account/cards/42/comments/c1")
      .with(
        body: { comment: { created_at: "2024-01-15T10:30:00Z" } }.to_json
      )
      .to_return(
        status: 200,
        body: '{"id": "c1", "created_at": "2024-01-15T10:30:00Z"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42", created_at: "2024-01-15T10:30:00Z" }).invoke(:update, ["c1"])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "2024-01-15T10:30:00Z", result["data"]["created_at"]
  end

  def test_delete_comment
    stub_request(:delete, "https://app.fizzy.do/test_account/cards/42/comments/c1")
      .to_return(status: 204, body: "")

    output = capture_output do
      Fizzy::Commands::Comment.new([], { card: "42" }).invoke(:delete, ["c1"])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert result["data"]["deleted"]
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

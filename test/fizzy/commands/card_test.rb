require "test_helper"

class Fizzy::Commands::CardTest < Fizzy::TestCase
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

  def test_list_returns_cards
    stub_request(:get, "https://app.fizzy.do/test_account/cards")
      .to_return(
        status: 200,
        body: '[{"id": "1", "number": 1, "title": "Card 1"}]',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Card.new.invoke(:list, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal 1, result["data"].length
    assert_equal "Card 1", result["data"][0]["title"]
  end

  def test_list_with_filters
    stub_request(:get, "https://app.fizzy.do/test_account/cards")
      .with(query: { "board_ids[]" => "10", "status" => "published" })
      .to_return(
        status: 200,
        body: '[{"id": "1", "title": "Filtered Card"}]',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Card.new([], { board: "10", status: "published" }).invoke(:list, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "Filtered Card", result["data"][0]["title"]
  end

  def test_show_returns_card
    stub_request(:get, "https://app.fizzy.do/test_account/cards/42")
      .to_return(
        status: 200,
        body: '{"id": "100", "number": 42, "title": "My Card", "description": "Details"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Card.new.invoke(:show, ["42"])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal 42, result["data"]["number"]
    assert_equal "My Card", result["data"]["title"]
  end

  def test_create_card
    stub_request(:post, "https://app.fizzy.do/test_account/boards/5/cards")
      .with(
        body: { card: { title: "New Card" } }.to_json,
        headers: { "Content-Type" => "application/json" }
      )
      .to_return(
        status: 201,
        body: '{"id": "200", "number": 50, "title": "New Card"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Card.new([], { board: "5", title: "New Card" }).invoke(:create, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "New Card", result["data"]["title"]
  end

  def test_create_with_description
    stub_request(:post, "https://app.fizzy.do/test_account/boards/5/cards")
      .with(
        body: { card: { title: "Card", description: "<p>Rich text</p>" } }.to_json
      )
      .to_return(
        status: 201,
        body: '{"id": "201", "title": "Card", "description": "<p>Rich text</p>"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Card.new([], { board: "5", title: "Card", description: "<p>Rich text</p>" }).invoke(:create, [])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "<p>Rich text</p>", result["data"]["description"]
  end

  def test_create_with_tags
    stub_request(:post, "https://app.fizzy.do/test_account/boards/5/cards")
      .with(
        body: { card: { title: "Card", tag_ids: ["1", "2"] } }.to_json
      )
      .to_return(
        status: 201,
        body: '{"id": "202", "title": "Card"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Card.new([], { board: "5", title: "Card", tag_ids: "1, 2" }).invoke(:create, [])
    end

    result = JSON.parse(output)
    assert result["success"]
  end

  def test_update_card
    stub_request(:put, "https://app.fizzy.do/test_account/cards/42")
      .with(
        body: { card: { title: "Updated Title" } }.to_json
      )
      .to_return(
        status: 200,
        body: '{"id": "100", "number": 42, "title": "Updated Title"}',
        headers: { "Content-Type" => "application/json" }
      )

    output = capture_output do
      Fizzy::Commands::Card.new([], { title: "Updated Title" }).invoke(:update, ["42"])
    end

    result = JSON.parse(output)
    assert result["success"]
    assert_equal "Updated Title", result["data"]["title"]
  end

  def test_delete_card
    stub_request(:delete, "https://app.fizzy.do/test_account/cards/42")
      .to_return(status: 204, body: "")

    output = capture_output do
      Fizzy::Commands::Card.new.invoke(:delete, ["42"])
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

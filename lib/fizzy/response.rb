module Fizzy
  class Response
    attr_reader :success, :data, :error, :pagination, :meta

    def self.success(data:, pagination: nil)
      new(success: true, data: data, pagination: pagination)
    end

    def self.error(code:, message:, status: nil, details: nil)
      new(
        success: false,
        error: {
          code: code,
          message: message,
          status: status,
          details: details
        }.compact
      )
    end

    def initialize(success:, data: nil, error: nil, pagination: nil)
      @success = success
      @data = data
      @error = error
      @pagination = pagination
      @meta = { timestamp: Time.now.utc.iso8601 }
    end

    def to_h
      result = { success: @success }
      result[:data] = @data if @data
      result[:error] = @error if @error
      result[:pagination] = @pagination if @pagination
      result[:meta] = @meta
      result
    end

    def to_json(*args)
      JSON.pretty_generate(to_h)
    end
  end
end

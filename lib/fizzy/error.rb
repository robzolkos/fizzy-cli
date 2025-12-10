module Fizzy
  # Exit codes
  EXIT_SUCCESS       = 0
  EXIT_ERROR         = 1
  EXIT_INVALID_ARGS  = 2
  EXIT_AUTH_FAILURE  = 3
  EXIT_FORBIDDEN     = 4
  EXIT_NOT_FOUND     = 5
  EXIT_VALIDATION    = 6
  EXIT_NETWORK       = 7

  class Error < StandardError
    attr_reader :exit_code

    def initialize(message, exit_code: EXIT_ERROR)
      super(message)
      @exit_code = exit_code
    end
  end

  class AuthError < Error
    def initialize(message = "Authentication failed")
      super(message, exit_code: EXIT_AUTH_FAILURE)
    end
  end

  class ForbiddenError < Error
    def initialize(message = "Permission denied")
      super(message, exit_code: EXIT_FORBIDDEN)
    end
  end

  class NotFoundError < Error
    def initialize(message = "Resource not found")
      super(message, exit_code: EXIT_NOT_FOUND)
    end
  end

  class ValidationError < Error
    def initialize(message = "Validation failed")
      super(message, exit_code: EXIT_VALIDATION)
    end
  end

  class NetworkError < Error
    def initialize(message = "Network error")
      super(message, exit_code: EXIT_NETWORK)
    end
  end

  class ConfigError < Error
    def initialize(message = "Configuration error")
      super(message, exit_code: EXIT_ERROR)
    end
  end
end

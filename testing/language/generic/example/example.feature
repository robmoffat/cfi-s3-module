Feature: Example usage of generic steps in Go
  As a developer
  I want to demonstrate the generic step patterns
  So that others can understand how to use them

  Background:
    Given I have an API client configured in "apiClient"
    And I have test data in "users"

  @example
  Scenario: Basic API testing with generic steps
    When I call "apiClient" with "Get" with parameter "/users"
    Then "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @example
  Scenario: Variable assignment and function calls
    Given "testValue" is a function which returns a value of "hello world"
    When I refer to "testValue" as "myFunction"
    And I call "myFunction"
    Then the function "result" should resolve
    And "{result}" is "hello world"

  @example
  Scenario: Counter functionality
    Given "clickHandler" is a invocation counter into "clickCount"
    When I call "clickHandler"
    And I call "clickHandler"
    And I call "clickHandler"
    Then "{clickCount}" is "3"

  @example
  Scenario: Array validation with data tables
    Then "{users}" is an slice of objects with the following contents
      | name     | active | profile.email    |
      | John Doe | true   | john@example.com |
      | Jane Doe | false  | jane@example.com |

  @example
  Scenario: Boolean and nil checks
    Given "trueValue" is a function which returns a value of "{true}"
    And "falseValue" is a function which returns a value of "{false}"
    And "nilValue" is a function which returns a value of "{nil}"
    When I call "{trueValue}"
    And I refer to "{result}" as "boolResult"
    Then "{boolResult}" is true
    When I call "{falseValue}"
    And I refer to "{result}" as "boolResult2"
    Then "{boolResult2}" is false
    When I call "{nilValue}"
    And I refer to "{result}" as "nilResult"
    Then "{nilResult}" is nil

  @example
  Scenario: Error handling
    Given "errorFunction" is a function which returns a value of "error: something went wrong"
    When I call "errorFunction"
    Then "{result}" is an error with message "error: something went wrong"

  @example
  Scenario: Timing operations
    Given we wait for a period of "100" ms
    Then "{result}" is nil

  @example
  Scenario: Object method calls
    When I call "apiClient" with "Get" with parameter "/status"
    Then "{result}" is not nil
    And "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

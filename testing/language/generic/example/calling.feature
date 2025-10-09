Feature: Function and method calling patterns
  As a developer
  I want to test function and method calls
  So that I can verify the calling patterns work correctly

  Background:
    Given I have an API client configured in "apiClient"
    And I have test data in "users"

  @calling
  Scenario: Basic API testing with generic steps
    When I call "{apiClient}" with "Get" with parameter "/users"
    Then "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @calling
  Scenario: Variable assignment and function calls
    Given "testValue" is a function which returns a value of "hello world"
    When I refer to "{testValue}" as "myFunction"
    And I call "{myFunction}"
    Then the function "result" should resolve
    And "{result}" is "hello world"

  @calling
  Scenario: Counter functionality
    Given "clickHandler" is a invocation counter into "clickCount"
    When I call "{clickHandler}"
    And I call "{clickHandler}"
    And I call "{clickHandler}"
    Then "{clickCount}" is "3"

  @calling
  Scenario: Object method calls
    When I call "{apiClient}" with "Get" with parameter "/status"
    Then "{result}" is not nil
    And "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

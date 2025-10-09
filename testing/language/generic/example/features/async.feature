Feature: Asynchronous operation patterns
  As a developer
  I want to test async operations
  So that I can verify goroutine-based task execution works correctly

  Background:
    Given I have an API client configured in "apiClient"

  @async
  Scenario: Start task by calling function
    Given "slowFunction" is a function which returns a value of "completed"
    When I start task "task1" by calling "{slowFunction}"
    Then I wait for task "task1" to complete
    And "{result}" is "completed"

  @async
  Scenario: Start task by calling function with parameter
    Given "addTen" is a function which adds 10 to a number
    When I start task "task2" by calling "{addTen}" with parameter "5"
    Then I wait for task "task2" to complete
    And "{result}" is "15"

  @async
  Scenario: Start task by calling function with two parameters
    Given "multiply" is a function which multiplies two numbers
    When I start task "task3" by calling "{multiply}" with parameters "6" and "7"
    Then I wait for task "task3" to complete
    And "{result}" is "42"

  @async
  Scenario: Start task by calling function with three parameters
    Given "concat" is a function which concatenates three strings
    When I start task "task4" by calling "{concat}" with parameters "Hello", " ", "World"
    Then I wait for task "task4" to complete
    And "{result}" is "Hello World"

  @async
  Scenario: Start task by calling object method
    When I start task "task5" by calling "{apiClient}" with "Get" with parameter "/data"
    Then I wait for task "task5" to complete
    And "{result}" is not nil

  @async
  Scenario: Start task by calling object method with parameter
    When I start task "task6" by calling "{apiClient}" with "Get" with parameter "/users"
    Then I wait for task "task6" to complete
    And "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @async
  Scenario: Wait for task with timeout
    Given "slowFunction" is a function which returns a value of "done"
    When I start task "task7" by calling "{slowFunction}"
    Then I wait for task "task7" to complete within "5000" ms
    And "{result}" is "done"

  @async
  Scenario: All-in-one wait for function
    Given "quickFunction" is a function which returns a value of "quick result"
    When I wait for "{quickFunction}"
    Then "{result}" is "quick result"

  @async
  Scenario: All-in-one wait for function with parameter
    Given "addTen" is a function which adds 10 to a number
    When I wait for "{addTen}" with parameter "20"
    Then "{result}" is "30"

  @async
  Scenario: All-in-one wait for function with two parameters
    Given "multiply" is a function which multiplies two numbers
    When I wait for "{multiply}" with parameters "8" and "9"
    Then "{result}" is "72"

  @async
  Scenario: All-in-one wait for function with three parameters
    Given "concat" is a function which concatenates three strings
    When I wait for "{concat}" with parameters "Go", "dog", "!"
    Then "{result}" is "Godog!"

  @async
  Scenario: All-in-one wait for object method
    When I wait for "{apiClient}" with "Get" with parameter "/data"
    Then "{result}" is not nil

  @async
  Scenario: All-in-one wait for object method with parameter
    When I wait for "{apiClient}" with "Get" with parameter "/status"
    Then "{result}" is an object with the following contents
      | status | message |
      |    200 | success |

  @async
  Scenario: Multiple concurrent tasks
    Given "task1Fn" is a function which returns a value of "result1"
    And "task2Fn" is a function which returns a value of "result2"
    And "task3Fn" is a function which returns a value of "result3"
    When I start task "concurrent1" by calling "{task1Fn}"
    And I start task "concurrent2" by calling "{task2Fn}"
    And I start task "concurrent3" by calling "{task3Fn}"
    When I wait for task "concurrent1" to complete
    Then "{result}" is "result1"
    When I wait for task "concurrent2" to complete
    Then "{result}" is "result2"
    When I wait for task "concurrent3" to complete
    Then "{result}" is "result3"

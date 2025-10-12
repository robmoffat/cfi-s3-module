@PerService @CCC.ObjStor @tlp-amber @tlp-red
Feature: CCC.ObjStor.C01.TR02
  As a security administrator
  I want to prevent any requests to read protected objects using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"

  Scenario: Service prevents reading object with untrusted KMS key
    Given I call "{api}" with "GetIAMService"
    And I refer to "{result}" as "iamService"
    And I call "{iamService}" with "ProvisionUser" with parameter "test-user-untrusted"
    And I refer to "{result}" as "testUser"
    And I call "{iamService}" with "SetAccess" with parameters "{testUser}", "{uid}" and "read"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUser}"
    And I refer to "{result}" as "userStorage"
    And I call "{storage}" with "CreateBucket" with parameter "test-bucket-obj-untrusted"
    And I refer to "{result}" as "bucket"
    And I call "{storage}" with "CreateObject" with parameters "{bucket.ID}", "test-object.txt" and "test content"
    When I call "{userStorage}" with "ReadObject" with parameters "{bucket.ID}" and "test-object.txt"
    Then "{result}" is an error
    And I call "{iamService}" with "DestroyUser" with parameter "{testUser}"
    And I call "{storage}" with "DeleteObject" with parameters "{bucket.ID}" and "test-object.txt"
    And I call "{storage}" with "DeleteBucket" with parameter "{bucket.ID}"

  Scenario: Service allows reading object with trusted KMS key
    Given I call "{api}" with "GetIAMService"
    And I refer to "{result}" as "iamService"
    And I call "{iamService}" with "ProvisionUser" with parameter "test-user-trusted"
    And I refer to "{result}" as "testUser"
    And I call "{iamService}" with "SetAccess" with parameters "{testUser}", "{uid}" and "read"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUser}"
    And I refer to "{result}" as "userStorage"
    And I call "{storage}" with "CreateBucket" with parameter "test-bucket-obj-trusted"
    And I refer to "{result}" as "bucket"
    And I call "{storage}" with "CreateObject" with parameters "{bucket.ID}", "test-object.txt" and "test content"
    When I call "{userStorage}" with "ReadObject" with parameters "{bucket.ID}" and "test-object.txt"
    Then "{result}" is not nil
    And I call "{iamService}" with "DestroyUser" with parameter "{testUser}"
    And I call "{storage}" with "DeleteObject" with parameters "{bucket.ID}" and "test-object.txt"
    And I call "{storage}" with "DeleteBucket" with parameter "{bucket.ID}"

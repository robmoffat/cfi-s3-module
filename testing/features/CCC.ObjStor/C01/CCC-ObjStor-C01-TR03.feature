@PerService @CCC.ObjStor @tlp-clear @tlp-green @tlp-amber @tlp-red
Feature: CCC.ObjStor.C01.TR03
  As a security administrator
  I want to prevent any requests to write to buckets using untrusted KMS keys
  So that data encryption integrity and availability are protected against unauthorized encryption

  Background:
    Given a cloud api for "{provider}" in "api"
    And I call "{api}" with "GetServiceAPI" with parameter "object-storage"
    And I refer to "{result}" as "storage"

  Scenario: Service prevents creating bucket with untrusted KMS key
    Given I call "{api}" with "GetIAMService"
    And I refer to "{result}" as "iamService"
    And I call "{iamService}" with "ProvisionUser" with parameter "test-user-untrusted"
    And I refer to "{result}" as "testUser"
    And I call "{iamService}" with "SetAccess" with parameters "{testUser}", "{uid}" and "write"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUser}"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateBucket" with parameter "test-bucket-write-untrusted"
    Then "{result}" is an error
    And I call "{iamService}" with "DestroyUser" with parameter "{testUser}"

  Scenario: Service allows creating bucket with trusted KMS key
    Given I call "{api}" with "GetIAMService"
    And I refer to "{result}" as "iamService"
    And I call "{iamService}" with "ProvisionUser" with parameter "test-user-trusted"
    And I refer to "{result}" as "testUser"
    And I call "{iamService}" with "SetAccess" with parameters "{testUser}", "{uid}" and "write"
    And I call "{api}" with "GetServiceAPIWithIdentity" with parameters "object-storage" and "{testUser}"
    And I refer to "{result}" as "userStorage"
    When I call "{userStorage}" with "CreateBucket" with parameter "test-bucket-write-trusted"
    Then "{result}" is not nil
    And "{result.ID}" is "test-bucket-write-trusted"
    And I call "{iamService}" with "DestroyUser" with parameter "{testUser}"
    And I call "{storage}" with "DeleteBucket" with parameter "{result.ID}"

# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [userspb/service.proto](#userspb_service-proto)
    - [CreateUserRequest](#users-CreateUserRequest)
    - [CreateUserResponse](#users-CreateUserResponse)
    - [DeleteUserRequest](#users-DeleteUserRequest)
    - [DeleteUserResponse](#users-DeleteUserResponse)
    - [GetAllUsersRequest](#users-GetAllUsersRequest)
    - [GetAllUsersResponse](#users-GetAllUsersResponse)
    - [GetUserRequest](#users-GetUserRequest)
    - [GetUserResponse](#users-GetUserResponse)
    - [UpdateUserRequest](#users-UpdateUserRequest)
    - [UpdateUserResponse](#users-UpdateUserResponse)
  
    - [AdminService](#users-AdminService)
    - [UserService](#users-UserService)
  
- [userspb/common.proto](#userspb_common-proto)
    - [UserData](#users-UserData)
  
- [Scalar Value Types](#scalar-value-types)



<a name="userspb_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## userspb/service.proto



<a name="users-CreateUserRequest"></a>

### CreateUserRequest
Create user request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |
| age | [int32](#int32) |  |  |






<a name="users-CreateUserResponse"></a>

### CreateUserResponse
Create user response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |
| age | [int32](#int32) |  |  |






<a name="users-DeleteUserRequest"></a>

### DeleteUserRequest
Delete user request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="users-DeleteUserResponse"></a>

### DeleteUserResponse
Delete user response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | [bool](#bool) |  |  |






<a name="users-GetAllUsersRequest"></a>

### GetAllUsersRequest
Get all users request






<a name="users-GetAllUsersResponse"></a>

### GetAllUsersResponse
Get all users response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [GetUserResponse](#users-GetUserResponse) | repeated |  |






<a name="users-GetUserRequest"></a>

### GetUserRequest
Get user request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="users-GetUserResponse"></a>

### GetUserResponse
Get user response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |
| age | [int32](#int32) |  |  |






<a name="users-UpdateUserRequest"></a>

### UpdateUserRequest
Update user request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |






<a name="users-UpdateUserResponse"></a>

### UpdateUserResponse
Update user response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |





 

 

 


<a name="users-AdminService"></a>

### AdminService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetAllUsers | [GetAllUsersRequest](#users-GetAllUsersRequest) | [GetAllUsersResponse](#users-GetAllUsersResponse) | Get all users (admin only) |


<a name="users-UserService"></a>

### UserService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateUser | [CreateUserRequest](#users-CreateUserRequest) | [CreateUserResponse](#users-CreateUserResponse) | Create a new user |
| GetUser | [GetUserRequest](#users-GetUserRequest) | [GetUserResponse](#users-GetUserResponse) | Get user by ID |
| UpdateUser | [UpdateUserRequest](#users-UpdateUserRequest) | [UpdateUserResponse](#users-UpdateUserResponse) | Update user |
| DeleteUser | [DeleteUserRequest](#users-DeleteUserRequest) | [DeleteUserResponse](#users-DeleteUserResponse) | Delete user |

 



<a name="userspb_common-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## userspb/common.proto



<a name="users-UserData"></a>

### UserData
User data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_uuid | [string](#string) |  | User UUID |
| name | [string](#string) |  | User name |
| last_name | [string](#string) |  | User last name |
| email | [string](#string) |  | User email |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time of creation |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time of update |
| last_login | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time of last login |





 

 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |


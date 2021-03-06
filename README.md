Requirements:

1. Listen on port[masked]. Handle the following commands: GET, SET, DEL, QUIT, BEGIN, COMMIT. All commands need to be terminated with a "\r\n" string.
3. Must support transactions. Transactions should be a N+1 queries grouped together to operate on the data, segmented off from other transactions.

Non-transactional Command Specifications:

1. GET - key -> returns ; GET takes a string key and returns the value if it exists otherwise an empty string.

2. SET - key, value -> returns ; SET takes a string for a key and a string for a value and returns a string. The return string will be "OK" on success or contain an error.

3. DEL - key; DEL takes a string for the key and removes that entry. The return string will be "OK" on success or contain an error.

4. QUIT - kills the active connection

Transactional Command Specifications:

1. BEGIN - BEGIN indicates to the server that you want to use a transaction and to start one. All commands following will be scoped to that transaction.

2. COMMIT - COMMIT indicates that the transaction is complete and to finalize the transaction by committing any changes to the global data store.

Examples:

Regular interaction with the server:

SET key1 value1\r\n
GET key1\r\n
DEL key1\r\n
QUIT\r\n

In the example above, the expectation is that if "key1" doens't exist, it is created and the "value1" is assinged to it. If it already exists, "key1" is updated with the new value.

Transactional interaction with the server:

BEGIN\r\n
SET key1 value1\r\n
GET key1\r\n
DEL key1\r\n
COMMIT\r\n
QUIT\r\n

n the example above, the expectation is that a new transaction is started which will cause all of the queries below to operate on their own set of data. That means that the same expectations apply from the first example however those changes aren't seen outside of the scope of the transaction. If "key1" is set, then deleted, another transaction could be performing operations on "key1" at the same time however only the transaction that gets committed is seen.

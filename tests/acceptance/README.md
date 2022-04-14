### API acceptance tests for Optimizely Agent

Acceptance tests run against a real Optimizely project, using REST API calls.
The project lives on app.optimzely.com and is maintained by the Full Stack team at Optimizely.

First, do everything from the agent's root directory. 

Python version 3.7 or greater is required.
It is recommended to set up a python virtual environment.     
Activate virtual environment.  
Install requirements `pip install -r tests/acceptance/requirements.txt` 


Run tests  
`MYHOST="http://localhost:8080" make test-acceptance`

You can point `MYHOST` to any URL where agent service is located.

If you want to run an individual test add TEST variable in front like so:
`TEST="test_activate__disable_tracking" MYHOST="http://localhost:8080" make test-acceptance`  
The TEST variable is based on Pytest's -k pattern matching flag so you can provide a full name of the test to only run that test, or a partial name which will run all tests that match that name pattern.

Tests have a function `url_points_to_cluster()` which has the role of toggling between Agent versions that support user profile service and those that don't.  
Optimizely runs Agent with these tests also on AWS clusters. UPS is not supported there. So we added this function  
so that Agent on AWS clusters excludes UPS tests. See PR [#341](https://github.com/optimizely/agent/pull/341).

When running all tests in one go with UPS they will pass. But running individual tests that require UPS may fail  (run all to be sure).

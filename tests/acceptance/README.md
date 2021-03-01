### API acceptance tests for Optimizely Agent

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

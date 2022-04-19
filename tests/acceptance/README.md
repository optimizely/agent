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

Tests contain a few tests that don't support user profile service. Those tests are intended to be used 
by Optimizely at a different place and are therefore excluded from the main test run.

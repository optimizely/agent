### API acceptance tests for Optimizely Agent

First, do everything from the agent's root directory. 

It is recommended to set up a python virtual environment (Py version 3.7+).    
Activate virtual environment.  
Install requirements `pip install -r tests/acceptance/requirements.txt` 


Run tests  
1. `pytest -v tests/acceptance/test_acceptance/ --host http://localhost:8080`

`--host` can point to any URL where agent service is located

To run an individual test  
TBD











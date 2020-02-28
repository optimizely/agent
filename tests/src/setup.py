import os

from setuptools import setup
from setuptools import find_packages

here = os.path.join(os.path.dirname(__file__))

with open(os.path.join(here, 'requirements.txt')) as _file:
    REQUIREMENTS = _file.read().splitlines()

setup(
    name="agent_acceptance_tests",
    version="0.1",
    author="Optimizely",
    description="Test framework to validate Agent API endpoints (acceptance tests).",
    license="",
    url="",
    packages=find_packages(),
    install_requires=REQUIREMENTS,
)

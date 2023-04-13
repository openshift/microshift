
from robot.api import logger, Failure
import re

# More examples:
# - https://docs.robotframework.org/docs/extending_robot_framework/custom-libraries/python_library
# - https://robotframework.org/robotframework/latest/RobotFrameworkUserGuide.html#creating-test-library-class-or-module

# Check HTTP Response
#     [Arguments]    ${result}
#     Log    ${result.stdout}
#     Log    ${result.stderr}
#     Should Be Equal As Integers    ${result.rc}    0
#     Should Match Regexp    ${result.stdout}    HTTP.*200
#     Should Match    ${result.stdout}    *Hello MicroShift*

def check_http_response(result):
    logger.info(result.stdout)
    logger.info(result.stderr)
    if result.rc != 0:
        raise Failure('Return code should be 0')
    if not re.match("HTTP.*200", result.stdout):
        raise Failure('not 200')
    if "Hello MicroShift" not in result.stdout:
        raise Failure('wrong body')

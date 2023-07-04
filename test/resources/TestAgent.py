import json
import libostree
from robot.libraries.BuiltIn import BuiltIn

_log = BuiltIn().log

CONFIG_PATH = "/var/lib/microshift-test-agent.json"

# Example config
# {
#     "deploy-id": {
#         "every": [ "prevent_backup" ],
#         "1": [ "fail_greenboot" ],
#         "2": [ "..." ],
#         "3": [ "..." ]
#     }
# }


class TestAgent:
    def __init__(self):
        self.cfg = dict()

    def add_action(self, deployment: str, boot: str, action: str) -> None:
        if deployment not in self.cfg:
            self.cfg[deployment] = {boot: [action]}

        elif boot not in self.cfg[deployment]:
            self.cfg[deployment][boot] = [action]

        elif action not in self.cfg[deployment][boot]:
            self.cfg[deployment][boot].append(action)

        _log(f"TestAgent Config: {self.cfg}")

    def add_action_for_next_deployment(self, boot: str, action: str) -> None:
        self.add_action("next", boot, action)

    def substitute_staged(self) -> None:
        if "next" not in self.cfg:
            _log("'next' deployment not found")
            return
        id = libostree.get_staged_deployment_id()
        self.cfg[id] = self.cfg["next"]
        del self.cfg["next"]

    def write(self) -> None:
        self.substitute_staged()
        j = json.dumps(self.cfg)
        BuiltIn().should_not_be_empty(j)
        libostree.remote_sudo(
            f"echo '{j}' | sudo tee /var/lib/microshift-test-agent.json"
        )

    def remove(self) -> None:
        self.cfg = {}
        libostree.remote_sudo("rm /var/lib/microshift-test-agent.json")

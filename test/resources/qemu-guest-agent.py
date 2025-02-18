"""
The qemu-guest-agent.py package uses virsh qemu-agent-command under the hood to interact with the guest-agent running on
a VM. Due to behaviours of the qemu-guest-agent, checking the status of a guest-agent's subprocess executes a
destructive read on success.  For that reason, keywords that assert the state of, or wait for, a guest process must also
return the status returned from the poll. Keywords assume a RHEL-like guest OS, which is apparent in hardcoded command
paths used here.

The guest-agent package provides the following
keywords:
    - get_guest_process_result
    - start_guest_process
    - wait_guest_for_process
    - run_guest_process
    - terminate_guest_process
    - read_from_file
    - write_to_file
    - guest_agent_is_ready

If you are looking for keywords to control the guest VM itself, see ./libvirt.resource
"""
from __future__ import annotations  # Support for Python 3.7 and earlier

import json
import argparse
from base64 import b64decode, b64encode
from os.path import basename

from robot.libraries.BuiltIn import BuiltIn, DotDict
from robot.libraries.Process import Process, ExecutionResult
from robot.utils.robottime import timestr_to_secs


def _execute(vm_name: str, agent_message: dict) -> dict | int:
    virsh_args = f'virsh --connect=qemu:///system qemu-agent-command --domain={vm_name} --cmd='
    msg = json.dumps(agent_message)
    virsh_args += f'\'{msg}\''

    result: ExecutionResult = Process().run_process(virsh_args, shell=True)
    # Only raise an error if the virsh command itself fails.  The guest-agent may return a non-zero exit code which
    # should be handled by the keyword caller.
    if result.rc != 0:
        raise RuntimeError(f'virsh command failed:\nstdout={result.stdout}'
                           f'\nstderr={result.stderr}'
                           f'\nrc={result.rc}')
    #  qemu-agent-command returns data to stdout as:
    #  {
    #    return: {
    #      pid: int
    #    }
    #  }
    return json.loads(result.stdout)['return']


def _do_guest_exec(vm_name: str, cmd: str, *args, env: dict, stdin: str) -> int:
    # _do_guest_exec wraps a given command and arg list into a qemu-guest-agent guest-exec API call.  For more info this
    # API, see https://qemu-project.gitlab.io/qemu/interop/qemu-ga-ref.html#qapidoc-211. For more information on the
    # guest-exec return message, see https://qemu-project.gitlab.io/qemu/interop/qemu-ga-ref.html#qapidoc-201
    agent_cmd_wrapper = {
        'execute': 'guest-exec',
        'arguments': {
            'path': cmd,
            'arg': args,
            'env': [f'{k}={v}' for k, v in env.items()] if env is not None else [],
            'input-data': b64encode(stdin.encode('utf-8')).decode('utf-8') if stdin is not None else '',
            'capture-output': True,
        }
    }
    #  _execute() returns "guest-exec" as dict containing only the PID of the command executed on the guest.
    # {
    #   pid: int
    # }
    content = _execute(vm_name, agent_cmd_wrapper)
    return content['pid']


def _do_guest_exec_status(vm_name: str, pid: int) -> (dict, bool):
    # _do_guest_exec_status wraps a given Process ID (pid) in a qemu-guest-agent guest-exec-status API call. For more
    # info this API, see https://qemu-project.gitlab.io/qemu/interop/qemu-ga-ref.html#qapidoc-198.  For information on
    # the guest-exec-status return message, see https://qemu-project.gitlab.io/qemu/interop/qemu-ga-ref.html#qapidoc-194
    agent_cmd_wrapper = {
        'execute': 'guest-exec-status',
        'arguments': {
            'pid': pid,
        }
    }
    #  qemu-agent-command "guest-exec" returns data to stdout as the example below. Fields marked (optional) will be
    #  undefined if exited is False. optional keys may not exist even if exited is True.
    #  {
    #    return: {
    #      exited: boolean
    #      exitcode: int (optional)
    #      signal: int (optional)
    #      out-data: string (optional)
    #      err-data: string (optional)
    #      out-truncated: boolean (optional)
    #      err-truncated: boolean (optional)
    #    }
    #  }
    content = _execute(vm_name, agent_cmd_wrapper)
    BuiltIn().log(f'guest-exec-status result: {content}')
    return {
        'rc': content['exitcode'] if 'exitcode' in content else None,
        'stdout': b64decode(content['out-data']).decode("utf-8").strip() if 'out-data' in content else '',
        'stderr': b64decode(content['err-data']).decode("utf-8").strip() if 'err-data' in content else '',
    }, content['exited']


def _do_kill(vm_name: str, pid: str, signal: str) -> int:
    if not vm_name:
        raise ValueError('vm name is not specified')
    if not pid:
        raise ValueError('pid is not specified')
    return _do_guest_exec(vm_name, '/usr/bin/kill', signal, pid)


def terminate_guest_process(vm_name: str, pid: str | int, kill: bool = False) -> (int, bool):
    """
    :param vm_name:         The name of the VM to execute the command on.
    :type vm_name:          str
    :param pid:             The process ID returned by qemu-agent-command.execute.guest-exec. This is the PID of the
                            command executed on the guest.
    :type pid:              str | int
    :param kill:            If True, the process is forcefully killed. Otherwise, the process is gracefully terminated.
    :type kill:             bool
    :return:                The exit code of the kill command and boolean ``exited``.
    :rtype:                 (int, bool)
    :raises RuntimeError:   If /bin/kill process returns a non-zero exit code.
    :raises ValueError:     If the PID is not an integer, or vm_name or pid is not specified.

    ``terminate_guest_process`` terminates or kills a process by its PID on the guest and returns the process's status.
    The process is terminated by sending a SIGTERM to the process. The status of the process is returned as a dict
    containing the stdout, stderr, and exit code of the process. If the process is still running (i.e. it ignored the
    SIGTERM) the `rc` value will be None and the `exited` value will be False. If the process has exited, the `rc` will
    be returned and `exited` will be True.

    Note: DOES NOT wrap the Process.terminate_process function. This is because the Process version only kills
    processes on the host, not the guest.

    Examples:
    ${stdout}    ${stderr}    ${rc}=    Terminate Guest Process    vm_name    pid
    Should Be Equal As Integers    ${rc}    -15
    Terminate Guest Process    vm_name    pid    kill=${True}
    """

    # pid is stored as an argument in the qemu-agent-command args list, which is expected to be a string
    if isinstance(pid, int):
        pid = str(pid)
    return _do_kill(vm_name, pid, "-15" if not kill else "-9")


def get_guest_process_result(vm_name: str, pid: int) -> (DotDict, bool):
    """
    :param vm_name:        The name of the VM to execute the command on
    :type vm_name:         str
    :param pid:            The process ID returned by qemu-agent-command.execute.guest-exec
    :type pid:             int
    :return:               DotDict with keys ``stdout``, ``stderr``, ``rc``, and boolean ``exited``.
    :rtype:                (DotDict, bool)
    :raises ValueError:    If the VM name or process ID is not specified
    :raises RuntimeError:  If guest-agent-command returns a non-zero exit code

    ``get_guest_process_result`` gets the results of the qemu-agent-command for a given process ID, which is returned by
    the qemu-agent-command.execute.guest-exec call immediately after execution. The PID is used to retrieve that
    process's stdout, stderr, and exit code. If the process is still running, the stdout, stderr, and exit code will be
    None and the boolean return value will be False. If the process has exited, the exit code will be returned and the
    boolean return value will be True.
    """
    if not pid:
        raise ValueError('PID is not specified')
    return DotDict(_do_guest_exec_status(vm_name, pid))


def wait_for_guest_process(vm_name: str, pid: int, timeout: int = None, on_timeout: str = "continue") -> (
        DotDict, bool):
    """
    :param vm_name:         The name of the VM to execute the command on
    :type vm_name:          str
    :param pid:             The process ID returned by qemu-agent-command.execute.guest-exec. This is the PID of the
                            command executed on the guest.
    :type pid:              int
    :param timeout:         The maximum amount of time to wait for the process to complete, in seconds.  If None,
                            wait indefinitely, (default).
    :type timeout:          int
    :param on_timeout:      The action to take if the timeout is reached.
                            documentation for more information. Default: "continue"
    :type on_timeout:       str
    :return:                DotDict with keys ``stdout``, ``stderr``, ``rc``, and boolean ``exited``.
    :rtype:                 (DotDict, bool)
    :raises RuntimeError:   If the VM name or process ID is not specified

    ``wait_for_guest_process`` is analogous to Process.wait_for_process(). The status of the command is retrieved using
    the qemu-agent-command.execute.guest-exec-status call.  If the process does not complete within the timeout, the
    ``on_timeout`` value will determine how to handle it.

    The ``on_timeout`` parameter can be one of the following values:
    | = Value = |               = Action =               |
    | continue  | The process is left running (default). |
    | terminate | The process is gracefully terminated.  |
    | kill      | The process is forcefully stopped.     |

    Examples:
    # Process ends cleanly
    ${stdout}    ${stderr}    ${rc}=    Wait For Guest Process    vm_name    pid
    Should Be Equal As Integers    ${rc}    0
    # Process does not end by timeout
    ${stdout}    ${stderr}    ${rc}    ${exited}=    Wait For Guest Process    vm_name    pid    timeout=10
    Should Not Be True    ${exited}
    # Kill process on timeout
    ${stdout}    ${stderr}    ${rc}    ${exited}=    Wait For Guest Process    vm_name    pid    timeout=10    on_timeout=kill
    Should Be True    ${exited}
    Should Be Equal as Integers    ${rc}    -9
    """
    if not vm_name:
        raise ValueError('vm name is not specified')
    if not pid:
        raise ValueError('pid is not specified')

    BuiltIn().log("waiting for process, timeout=%s, on_timeout=%s" % (timeout, on_timeout))
    timeout = timestr_to_secs(timeout) if timeout is not None else None
    status, exited = _do_guest_exec_status(vm_name, pid)
    while not exited:
        if timeout is not None:
            timeout -= 1
            if timeout <= 0:
                if on_timeout == 'terminate':
                    BuiltIn().log(f'process {pid} on {vm_name} timed out, terminating', 'WARN')
                    terminate_guest_process(vm_name, pid)
                elif on_timeout == 'kill':
                    BuiltIn().log(f'process {pid} on {vm_name} timed out, killing', 'WARN')
                    terminate_guest_process(vm_name, pid, kill=True)
                elif on_timeout == 'continue':
                    BuiltIn().log(f'process {pid} on {vm_name} timed out, continuing', 'WARN')
                    break
                else:
                    raise ValueError(f'invalid on_timeout value: {on_timeout}')
                status, exited = _do_guest_exec_status(vm_name, pid)
                break
        BuiltIn().sleep(1)
        status, exited = _do_guest_exec_status(vm_name, pid)

    BuiltIn().log(f'process {pid} on {vm_name} exited, returning result: {status}')
    return DotDict(status), exited


def run_guest_process(vm_name: str, cmd: str, *args, env: dict = None, stdin: str = None, timeout: int = None,
                      on_timeout: str = "continue") -> (DotDict, bool):
    """
    :param vm_name:     The name of the VM to execute the command on
    :type vm_name:      str
    :param cmd:         The absolute path to a command on the guest, e.g. "/bin/ls"
    :type cmd:          str
    :param args:        The arguments to pass to the command, separated into a list of strings, e.g. ["-l", "/tmp"]
                        Command arguments which take a value must be passed as a single string, e.g. ["-f /tmp""]
    :type args:         list
    :param env:         A dictionary of environment variables to set for the command, e.g. {"PATH": "/bin:/usr/bin"}
    :type env:          dict
    :param stdin:       The input to pass to the command.  If None, no input is passed to the command. (default: None)
    :type stdin:        str
    :param timeout:     The maximum amount of time to wait for the process to complete, in seconds.  If None,
                        wait indefinitely (default).
    :type timeout:      int
    :param on_timeout:  The action to take if the timeout is reached.
    :type on_timeout:   str
    :return:            The stdout, stderr, and exit code of the guest process for the given PID
    :rtype:             (DotDict, bool)
    :raises             RuntimeError: If the VM name or command path is not specified

    ``run_guest_process`` excepts a given command and optional arg list to execute on a given VM using the
    qemu-agent-command. The command is executed by using virsh qemu-guest-command. This is a blocking call unless
    qemu-agent-timeout is set to anything but the default (must be configured externally with virsh, see
    https://www.libvirt.org/manpages/virsh.html#guest-agent-timeout).

    Usage: ${stdout}    ${stderr}    ${rc}=    Run Guest Process    vm_name    cmd    *args
    Examples:
    ${stdout}    ${stderr}    ${rc}=    Run Guest Process    vm-host-1    /bin/ls    -l    /tmp    --color=never
    Log Many   ${stdout}    ${stderr}
    Should Be Equal As Integers    ${rc}    0

    ${stdout}    ${stderr}    ${rc}    ${exited}=    Run Guest Process    vm-host-1    /bin/sleep    30s   timeout=10
    Should Be True    ${exited}
    """

    if not vm_name:
        raise ValueError('vm name is not specified')
    if not cmd:
        raise ValueError('cmd is not specified')

    pid = _do_guest_exec(vm_name, cmd, *args, env=env, stdin=stdin)
    return wait_for_guest_process(vm_name, pid, timeout, on_timeout)


def start_guest_process(vm_name: str, cmd: str, *args, env: dict, stdin: str = None) -> int:
    """
    :param vm_name:         The name of the VM to execute the command on
    :type vm_name:          str
    :param cmd:             The command to execute on the guest, e.g. "/bin/ls".  Must be a path.
    :type cmd:              str
    :param args:            The arguments to pass to the command, separated into a list of strings, e.g. ["-l", "/tmp"]
                            Command arguments which take a value must be passed as a single string, e.g. ["-f /tmp""]
    :type args:             list
    :param env:             A dictionary of environment variables to set for the command, e.g. {"PATH": "/bin:/usr/bin"}
    :type env:              dict
    :param stdin:           The input to pass to the command.  If None, no input is passed to the command. (default: None)
    :type stdin:            str
    :return:                The guest's process ID of the command
    :rtype:                 int | None
    :raises ValueError      If the VM name or command path is not specified
    :raises RuntimeError:   If the qemu-agent-command returns a non-zero exit code

    ``start_guest_process`` is analogous to Process.start_process in that it does not wait for a guest process to exit.
    The command is executed on a VM using the qemu-agent-command.  To retrieve the guest process's status, use the
    PID returned by start_guest_process with get_guest_process_result.

    Examples:
    ${pid}=    Start Guest Process    vm-host-1    /bin/ls    -l    /tmp
    ${stdout}    ${stderr}    ${rc}=    Get Guest Process Result    vm-host-1    ${pid}
    Log Many    ${stdout}    ${stderr}
    Should Be Equal As Integers    ${rc}    0
    """
    if not vm_name:
        raise ValueError('vm name is not specified')
    if not cmd:
        raise ValueError('cmd is not specified')
    return _do_guest_exec(vm_name, cmd, *args, env=env, stdin=stdin)


def _open_file(vm_name: str, path: str, mode: str) -> int:
    agent_cmd_wrapper = {
        'execute': 'guest-file-open',
        'arguments': {
            'path': path,
            'mode': mode,
        }
    }
    # guest-file-open returns an integer handle to the file opened on the guest
    return _execute(vm_name, agent_cmd_wrapper)


def _close_file(vm_name, handle):
    agent_cmd_wrapper = {
        'execute': 'guest-file-close',
        'arguments': {
            'handle': handle,
        }
    }
    # guest-file-close returns nothing on success, raises an error on failure
    _execute(vm_name, agent_cmd_wrapper)


def read_from_file(vm_name: str, path: str) -> str:
    """
    :param vm_name:         The name of the VM to execute the command on
    :type vm_name:          str
    :param path:            The absolute path to a file on the guest, e.g. "/tmp/foo"
    :type path:             str
    :return:                The contents of the file
    :rtype:                 str
    :raises ValueError:     If the VM name or file path is not specified
    :raises RuntimeError:   If the qemu-agent-command returns a non-zero exit code
    :raises RuntimeError:   If the file cannot be opened on the guest

    ``qemu-guest-read`` reads from a file on the guest VM. The file is opened using the qemu-agent-command
    guest-file-open API call, and read using the guest-file-read API call.  The file is unconditionally closed using the
    guest-file-close API call. See  https://qemu-project.gitlab.io/qemu/interop/qemu-ga-ref.html#qapidoc-54 for more
    information on the guest-file-read API call. See https://qemu-project.gitlab.io/qemu/interop/qemu-ga-ref.html#qapidoc-54
    for more information on the guest-file-close API call.
    """
    handle = _open_file(vm_name, path, 'r')
    try:
        content = _execute(vm_name, {
            'execute': 'guest-file-read',
            'arguments': {
                'handle': handle,
                'count': 40960
            }
        })
    finally:
        _close_file(vm_name, handle)

    return b64decode(content['buf-b64']).decode('utf-8')


def write_to_file(vm_name: str, path: str, content: str, append=False) -> int:
    """
    :param vm_name:         The name of the VM to execute the command on
    :type vm_name:          str
    :param path:            The absolute path to a file on the guest, e.g. "/tmp/foo"
    :type path:             str
    :param content:         The contents to write to the file
    :type content:          str
    :param append:          If True, append to the file. Otherwise, overwrite the file. (default: False)
    :type append:           bool
    :return:                The number of bytes written to the file
    :rtype:                 int
    :raises ValueError:     If the VM name or file path is not specified
    :raises RuntimeError:   If the qemu-agent-command returns a non-zero exit code.

    ``write_to_file`` writes to a file on the guest VM. The file is opened using the qemu-agent-command guest-file-open
    API call, and written to using the guest-file-write API call.  The file is unconditionally closed using the
    guest-file-close API call. See https://qemu-project.gitlab.io/qemu/interop/qemu-ga-ref.html#qapidoc-54 for more
    information on the guest-file-write API call.
    """
    if not vm_name:
        raise ValueError('vm name is not specified')
    if not path:
        raise ValueError('path is not specified')
    handle = _open_file(vm_name, path, mode='a' if append else 'w')
    try:
        content = _execute(vm_name, {
            'execute': 'guest-file-write',
            'arguments': {
                'handle': handle,
                'buf-b64': b64encode(content.encode('utf-8')).decode('utf-8'),
            }
        })
    finally:
        _close_file(vm_name, handle)
    # ``guest-file-write`` returns the number of bytes written to the file and boolean 'eof' if the end of file was
    # encountered while writing.  Only return the number of bytes written. Returning "EOF" on a write command is not
    # useful information.
    return content['count']


def guest_agent_is_ready(vm_name: str):
    """
    :param vm_name:         The name of the VM to execute the command on
    :type vm_name:          str
    :raises ValueError:     If the VM name is not specified
    :raises RuntimeError:   If the guest-ping fails or qemu-guest-command returns a non-zero exit code
    """
    if vm_name is None:
        raise ValueError('vm name is not specified')
    _execute(vm_name, {
        'execute': 'guest-ping',
    })

def _find_files(vm_name: str, dir: str, pattern: str):
    content, _ = run_guest_process(vm_name, "/bin/find", dir, "-maxdepth", "1", "-name", pattern)
    return content['stdout']

def _download_file(vm_name: str, src: str, dst: str):
    print(f"Downloading {src}")

    handle = _open_file(vm_name, src, 'r')
    try:
        content = _execute(vm_name, {
            'execute': 'guest-file-read',
            'arguments': {
                'handle': handle,
                'count': 40960
            }
        })
    finally:
        _close_file(vm_name, handle)

    content = b64decode(content['buf-b64'])

    with open(dst, "wb") as f:
        f.write(content)

def download_files(vm_name: str, src_dir: str, dst_dir: str, pattern: str):
    """
    :param vm_name:     The name of the VM to download the files from
    :type vm_name:      str
    :param src_dir:     Source directory where the files are located on the VM
    :type src_dir:      str
    :param dst_dir:     Destination directory where the downloaded files should be saved
    :type dst_dir:      str
    :param pattern:     Pattern or filename, e.g. sosreport-*, journal.log
    :type pattern:      str
    """
    if (not src_dir.endswith('/')):
        src_dir = src_dir + '/'
    if (not dst_dir.endswith('/')):
        dst_dir = dst_dir + '/'

    files = _find_files(vm_name, src_dir, pattern).splitlines()
    for file in files:
        filename = basename(file)
        src = src_dir + filename
        dst = dst_dir + filename
        _download_file(vm_name, src, dst)
    

def upload_file(vm_name: str, src: str, dst: str):
    """
    :param vm_name:     The name of the VM to upload the file to
    :type vm_name:      str
    :param src:         The absolute path to a source local file, e.g. "/tmp/foo"
    :type src:          str
    :param dst:         The absolute path to the destination file on guest, e.g. "/tmp/foo"
    :type dst:          str
    """
    print(f"Uploading {src}")

    with open(src, "r") as f:
        content = f.read()
    
    write_to_file(vm_name, dst, content)

def main():
    parser = argparse.ArgumentParser()

    subparsers = parser.add_subparsers(dest="command", required=True)

    cp_from_parser = subparsers.add_parser("download", help="Copy file(s) from VM")
    cp_from_parser.add_argument("--vm", required=True, help="domain name")
    cp_from_parser.add_argument("--src_dir", required=True, help="source path on VM")
    cp_from_parser.add_argument("--dst_dir", required=True, help="local destination path")
    cp_from_parser.add_argument("--pat", required=True, help="filename or pattern")

    cp_to_parser = subparsers.add_parser("upload", help="Copy file to VM")
    cp_to_parser.add_argument("--vm", required=True, help="domain name")
    cp_to_parser.add_argument("--src", required=True, help="path to local file")
    cp_to_parser.add_argument("--dst", required=True, help="destination path on VM")

    bash_parser = subparsers.add_parser("bash", help="run bash command on VM")
    bash_parser.add_argument("--vm", required=True, help="domain name")
    bash_parser.add_argument("--args", required=True, help="arguments")

    args = parser.parse_args()

    if args.command == "download":
        download_files(args.vm, args.src_dir, args.dst_dir, args.pat)
    elif args.command == "upload":
        upload_file(args.vm, args.src, args.dst)
    elif args.command == "bash":
        print(f"Running {args.args}")
        run_guest_process(args.vm, "/bin/bash", "-c", args.args)

if __name__ == "__main__":
    main()
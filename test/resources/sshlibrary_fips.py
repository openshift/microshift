import paramiko
paramiko.Transport._preferred_kex = ('ecdh-sha2-nistp256','ecdh-sha2-nistp384')
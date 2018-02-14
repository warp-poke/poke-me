# Poke-me

Poke-me help you to sync a Git repo and servers file systems.
You have to register a Git repository in poke-me and a path.

## Depedancies

* Zookeeper

## Configuration

```yaml
listen: 127.0.0.1:8081
zk.servers: 127.0.0.1
cloner:
  ssh.key: /keys/key.private
  git:
    url: git@github.com:star-projects/assets.git
    secret: xxxxxxxxxxxxxxxx
  path: /opt/assets/
secrets:
    my-password: 1234
```
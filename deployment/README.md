# podops-infra
Automation of the podops.dev infrastructure

Create a ssh key:

```shell
ssh-keygen -t rsa -b 4096 -C "your_email@example.com" -f ~/.ssh/podops
```

Get the public key:

```shell
cat ~/.ssh/podops.pub
```

### Installation

Make a copy of `inventory/inventory.example.yml` and change values according to your needs.

To deploy e.g. a self-contained podops instance, use the `prepare.yml` playbook:

```shell
ansible-playbook -i inventory/<your_inventory> playbooks/prepare_cdn.yml
```

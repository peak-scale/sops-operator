export VAULT_ADDR=http://openbao.openbao.svc.cluster.local:8200
export VAULT_TOKEN=root

bao secrets enable -path=sops transit || true

bao write -force sops/keys/key-1
bao write -force sops/keys/key-2

sops -e secret-key-1.yaml > secret-key-1.enc.yaml
sops -e secret-key-2.yaml > secret-key-2.enc.yaml
sops -e secret-multi.yaml > secret-multi.enc.yaml
sops -e secret-quorum.yaml > secret-quorum.enc.yaml

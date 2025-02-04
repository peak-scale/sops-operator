package decryptor

import "fmt"

type MissingKubernetesSecret struct {
	Secret    string
	Namespace string
}

func (e *MissingKubernetesSecret) Error() string {
	return fmt.Sprintf("Secret not found: %s/%s", e.Namespace, e.Secret)
}

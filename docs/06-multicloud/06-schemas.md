# Schémas

Quatre vues, générées depuis le code avec noodle. Pour les régénérer : [`architecture/README.md`](https://github.com/3-Istor/cnp-docs/blob/main/architecture/README.md). Pour agir : [procédures opérateur](07-procedures-operateur.md).

## Où tourne quoi

*Ce qui reste sur le hub on-prem, ce qui part avec chaque projet vers son cloud cible, et quels liens traversent.*

![CNP multicloud](../01-architecture/diagrams/multicloud.light.png#only-light)
![CNP multicloud](../01-architecture/diagrams/multicloud.dark.png#only-dark)

- Chemin de requête (plan de données) : l'utilisateur atteint l'app par un tunnel Cloudflare sortant, sur le cluster cible.
- Plan de contrôle (arrière-plan) : Argo CD, Vault, CMP restent sur le hub.
- La carte *Gaps* liste l'écart entre le plan et le code ; les badges `Gx` y renvoient.

## Liens hub / clouds

*Quels liens traversent entre le hub et les clouds publics, qui les ouvre, et qu'est-ce qui les autorise ?*

![CNP multicloud links](../01-architecture/diagrams/multicloud-links.light.png#only-light)
![CNP multicloud links](../01-architecture/diagrams/multicloud-links.dark.png#only-dark)

- Chaque flèche part de l'initiateur : WireGuard est ouvert depuis l'on-prem (UDP 31820), jamais depuis le cloud.
- Argo CD, Vault et le trafic applicatif passent aujourd'hui par Internet ; le VPN ne porte que des flux de test.
- Les API des clusters ne répondent qu'aux CIDR autorisés (`api_allowed_cidrs`).

## Qui peut atteindre un projet

*Qu'est-ce qui peut atteindre les workloads d'un projet sur chaque cluster, et qu'est-ce qui l'impose ?*

![CNP multicloud policies](../01-architecture/diagrams/multicloud-policies.light.png#only-light)
![CNP multicloud policies](../01-architecture/diagrams/multicloud-policies.dark.png#only-dark)

- Par cluster : CiliumClusterwideNetworkPolicy on-prem, NetworkPolicy sur EKS et GKE, mêmes sources autorisées.
- Garde-fous GitOps : les AppProjects bornent namespaces, destinations et objets cluster.
- Aucun LoadBalancer : pas de chemin entrant vers les workloads.

## `cnp.sh up`

*Que se passe-t-il, dans l'ordre, quand l'opérateur lance `cnp.sh up` pour un cloud public ?*

![CNP cnp.sh up](../01-architecture/diagrams/cloud-up.light.png#only-light)
![CNP cnp.sh up](../01-architecture/diagrams/cloud-up.dark.png#only-dark)

- Lire dans l'ordre : Terraform (cluster, nœud, addons), puis ce qui est automatique (Argo, Vault, WireGuard).
- Les étapes d'`up` posent les secrets de liaison : jeton Argo, config Vault, clé WireGuard, endpoint.
- Ce qui survit à `cnp.sh down` est listé dans les [procédures opérateur](07-procedures-operateur.md#down).

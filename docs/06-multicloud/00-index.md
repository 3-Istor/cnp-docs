# Chantier multicloud — porter la couche projet/application sur on-prem, AWS et GCP

Ce chapitre est le dossier de référence du chantier multicloud. Il fige ce qui a été
décidé, ce qui est volontairement hors périmètre, et ce que la plateforme doit prouver
avant d'être considérée comme livrée.

## Contrainte non négociable

Une application déjà déployée continue de servir, d'authentifier ses utilisateurs et de
rafraîchir leurs jetons **même si le cluster on-prem devient injoignable**. Déployer et
provisionner peuvent s'arrêter ; servir, non.

## Documents

| Document | Ce qu'il contient |
| --- | --- |
| [Invariants et périmètre](01-invariants-et-perimetre.md) | I-1 à I-6, non-goals, définition de « terminé » |
| [Décisions D-01 à D-13](02-decisions.md) | Chaque décision, son alternative écartée, et ce qu'elle coûte |
| [Contrat d'interface avec l'équipe cluster](03-contrat-equipe-cluster.md) | Ce que CNP possède, ce que l'équipe cluster possède, les flux autorisés sur le VPN |
| [Chantiers et séquencement](04-chantiers-et-sequencement.md) | WS-0 à WS-9, phases, portes de passage |
| [Nettoyage et décommissionnement](05-nettoyage-et-decommissionnement.md) | Le job de clean par projet et par cloud |

## État

Décisions signées le 2026-09-15. Les implémentations correspondantes existent en
pull requests *draft* sur les dépôts concernés ; **rien n'a été appliqué** — ni
`terraform apply`, ni synchronisation Argo CD.

## Vocabulaire

| Terme | Sens dans ce chapitre |
| --- | --- |
| **Plan de contrôle** | Ce qui vit exclusivement on-prem : CMP, Argo CD, le runner Terraform, le Vault maître, le Keycloak d'administration |
| **Plan de données** | Les clusters cibles (on-prem, AWS, GCP) qui exécutent les charges des projets |
| **Cible / target cloud** | Le cloud choisi à la création d'un projet, immuable ensuite |
| **Registre** | `cnp-projects/registry/projects/<nom>.yaml`, la source de vérité Git du placement |

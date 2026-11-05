# Contrats partagés

Ce dossier contient les contrats communs aux pistes. Statut de chaque élément :

- **Gelés et testés** : `actions.json` (les actions nommées, avec leur classe de confidentialité) et `consent-scopes.json` (les scopes de consentement par personne). Les tests de ce paquet les gardent alignés avec le coffre (`intimacy.KnownScopes()`) et refusent les noms d'action dangereux (`payment.*`, noms hors du format `^[a-z]+(\.[a-z_]+)+$`).
- **Périmé** : `openapi/local.json` ne décrit plus l'API réelle du Core. Le code du Core est la source de vérité jusqu'à ce que la piste B le rafraîchisse. Routes manquantes : `/v1/privacy/{blind,open-eyes,status,leases,leases/validate}`, `/v1/subjects*`, `/v1/intimacy/library/*`, `/v1/intimacy/context`, `/v1/intimacy/resume`. L'ancien `openapi/native-v1.json` (API Rust abandonnée) a été supprimé.
- **Différé** : le contrat des événements de confidentialité (privacy-events) attend qu'une piste consomme une entrée perceptive. Aucune ne le fait pour l'instant.

`actions.json` est la liste d'autorisation du Policy Broker : il ne change qu'après relecture par la piste B ou par le contrôleur.

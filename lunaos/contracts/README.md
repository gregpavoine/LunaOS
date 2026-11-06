# Contrats partagés

Ce dossier contient les contrats communs aux pistes. Statut de chaque élément :

- **Gelés et testés** : `actions.json` (les actions nommées, avec leur classe de confidentialité) et `consent-scopes.json` (les scopes de consentement par personne). Les tests de ce paquet les gardent alignés avec le coffre (`intimacy.KnownScopes()`) et refusent les noms d'action dangereux (`payment.*`, noms hors du format `^[a-z]+(\.[a-z_]+)+$`).
- **Rafraîchi** : `openapi/local.json` décrit les routes du Core, y compris `/v1/privacy/*`, `/v1/subjects*`, `/v1/intimacy/library/*`, `/v1/intimacy/context` et `/v1/intimacy/resume`. Un test du Core vérifie que chaque route enregistrée y figure. L'ancien `openapi/native-v1.json` (API Rust abandonnée) a été supprimé.
- **Différé** : le contrat des événements de confidentialité (privacy-events) attend qu'une piste consomme une entrée perceptive. Aucune ne le fait pour l'instant.

`actions.json` est la liste d'autorisation du Policy Broker : il ne change qu'après relecture par la piste B ou par le contrôleur.

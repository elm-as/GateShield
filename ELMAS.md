# ELMAS.md — Charte de développement ElmasCore

> Ce fichier est la signature d'un projet ElmasCore. Il définit les règles non négociables
> à respecter par tout humain ou tout agent IA qui contribue à ce dépôt. Il doit être lu
> avant toute génération de code, toute refonte, tout ajout de dépendance.
>
> "From Data to Systems" — un projet ElmasCore se reconnaît à sa rigueur, pas à sa rapidité d'exécution.

---

## 1. Architecture

- L'architecture existante du projet fait autorité. On ne la contourne pas, on ne la
  duplique pas en parallèle, on ne crée pas un second pattern "plus simple" à côté de
  celui en place.
- Toute modification structurelle (nouveau dossier racine, changement de convention de
  nommage, nouvelle couche) doit être justifiée en une phrase avant d'être appliquée, pas
  après.
- Séparation stricte des responsabilités : UI, logique métier, accès aux données et
  configuration ne vivent jamais dans le même fichier. Un composant qui fait une requête
  réseau, calcule une règle métier et gère son propre style est un composant à corriger,
  pas un raccourci acceptable.
- Pas de dépendance ajoutée "parce que c'est plus rapide". Toute nouvelle librairie doit
  répondre à un besoin qu'aucune brique déjà présente ne couvre.

## 2. Taille et complexité du code

- **350 lignes maximum par fichier**, sans exception de confort. Un fichier qui approche
  cette limite est un signal : il fait probablement plusieurs choses et doit être découpé
  avant d'être étendu, pas après.
- Une fonction, une responsabilité. Si une fonction a besoin d'un commentaire pour expliquer
  "d'abord on fait X, puis Y, puis Z", c'est qu'elle doit devenir trois fonctions.
- Pas de duplication silencieuse : si une logique existe déjà ailleurs dans le projet, on
  la réutilise ou on l'extrait, on ne la recopie pas.

## 3. Choix technique

- On choisit **la meilleure stack pour le problème posé, jamais la plus rapide à sortir**.
  React par défaut parce que "c'est ce que toutes les IA proposent" n'est pas un choix,
  c'est un réflexe — et un réflexe n'est jamais une justification technique valable dans
  un projet ElmasCore.
- Le choix de stack se justifie par : la nature du produit (desktop, mobile, temps réel,
  data-intensif, offline-first…), la cible réelle, la maintenabilité à 2 ans, et le coût
  d'exploitation — jamais par la familiarité de l'outil qui génère le code.
- Avant d'écrire la première ligne, on nomme explicitement pourquoi cette stack et pas une
  autre. Si la réponse est "parce que c'est le plus rapide à mettre en place", on
  recommence la réflexion.

## 4. UI / UX

- Aucune interface ElmasCore ne doit ressembler à un template SaaS générique généré par
  IA : palette indigo/violet par défaut, cartes blanches arrondies avec icône-dans-carré,
  pills flottantes partout, dégradés décoratifs. Si l'interface pourrait appartenir à
  n'importe quel autre produit, elle est à refaire.
- Chaque projet a sa propre identité visuelle, ancrée dans son sujet réel (son
  utilisateur, son métier, son vocabulaire), pas dans un kit de composants par défaut.
  Palette, typographie et structure se décident pour ce produit précis.
- Densité et lisibilité priment sur la décoration. Un outil de gestion utilisé tous les
  jours n'a pas besoin d'ombres portées, de radius excessifs ou d'animations
  d'atterrissage — il a besoin d'être clair en une seconde.
- Chiffres, montants, identifiants : toujours en police à chiffres tabulaires, alignés
  correctement. Ce n'est pas un détail, c'est un signe de sérieux.

## 5. Qualité et maintenabilité

- Nommage explicite en toute circonstance : pas d'abréviations obscures, pas de `data2`,
  `temp`, `handleStuff`.
- Toute erreur est gérée explicitement, jamais avalée silencieusement.
- Le code doit pouvoir être repris par quelqu'un d'autre que son auteur sans réunion
  d'explication préalable.

## 6. Sécurité et données

- Aucune donnée sensible (clé API, identifiant, mot de passe) en dur dans le code ou
  versionnée. Toujours via variables d'environnement ou secret manager.
- Toute donnée utilisateur (élève, client, transaction) est traitée comme sensible par
  défaut, même en phase de prototype.

## 7. Documentation et livraison

- Un README à jour à la racine : ce que fait le projet, comment le lancer, comment
  contribuer. Un projet sans README à jour n'est pas considéré comme livrable.
- Les décisions structurantes (choix de stack, changement d'architecture) sont
  consignées, pas seulement discutées dans un chat qui disparaît.

---

*Ces règles s'appliquent à tout agent IA (Claude Code, Cursor, etc.) travaillant sur ce
dépôt, au même titre qu'à tout contributeur humain. En cas de conflit entre une
instruction ponctuelle et ce fichier, ce fichier fait foi sauf mention explicite du
contraire par le porteur du projet.*

— **Elmas**, ElmasCore

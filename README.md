[(Français)](#contr%C3%B4leur-bloc-note-prob)

## Prob Notebook Controller

Based on https://github.com/StatCan/kubeflow-controller
**There is a dependency on that repository** as well since it needs to be imported in order to use the `Notebook` struct and `NotebookInformer`.

This controller creates and deletes Authorization Policies based on Notebooks. 
The Authorization Policies block uploads and downloads on jupyterlab and rstudio images.

### How to use
Go check the [README](https://github.com/StatCan/profile-state-controller/blob/master/kind/README.md) in the kind folder for instruction on running it locally.

### How to Contribute

See [CONTRIBUTING.md](CONTRIBUTING.md)

### License

Unless otherwise noted, the source code of this project is covered under Crown Copyright, Government of Canada, and is distributed under the [MIT License](LICENSE).

The Canada wordmark and related graphics associated with this distribution are protected under trademark law and copyright law. 
No permission is granted to use them outside the parameters of the Government of Canada's corporate identity program. 
For more information, see [Federal identity requirements](https://www.canada.ca/en/treasury-board-secretariat/topics/government-communications/federal-identity-requirements.html).


## Contrôleur bloc-note prob 

Basé sur https://github.com/StatCan/kubeflow-controller
**Il y a une dépendance sur ce répertoire** puique certains de ses éléments sont importés de façon a utilisé la structure `Notebook` et `NotebookInformer`.

Ce controller créer et supprime les Authorization Policies basées sur les Noteboks.
Les Authorization Pol;icies bloquent les téléchargement ou téléversment de sur les image de jupyterlab et rstudio.

### Comment utiliser
Aller voir le [README](https://github.com/StatCan/profile-state-controller/blob/master/kind/README.md) dans le dossier kind pour les instructions pour l'exécution locale.

### Comment contribuer

Voir [CONTRIBUTING.md](CONTRIBUTING.md)

### Licence

Sauf indication contraire, le code source de ce projet est protégé par le droit d'auteur de la Couronne du gouvernement du Canada et distribué sous la [licence MIT](LICENSE).

Le mot-symbole « Canada » et les éléments graphiques connexes liés à cette distribution sont protégés en vertu des lois portant sur les marques de commerce et le droit d'auteur. 
Aucune autorisation n'est accordée pour leur utilisation à l'extérieur des paramètres du programme de coordination de l'image de marque du gouvernement du Canada. 
Pour obtenir davantage de renseignements à ce sujet, veuillez consulter les [Exigences pour l'image de marque](https://www.canada.ca/fr/secretariat-conseil-tresor/sujets/communications-gouvernementales/exigences-image-marque.html).

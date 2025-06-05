# TODO LIST

## Ansible oder Azure Keyvault fuer smtp daten , ssh keys

## SPF Record , PTR Record ( muss auf den Load Balancer zeigen)

## Die ganze Infra nur noch per Bastion(Jump Host zugaenglich machen)

### Load Balancer bauen oder einkaufen, Message Broker als Cache fuer die Emails im Falle das Postfix VM down geht

#### SPAM Filter Service fuer Postgres Docker Image bauen, separater Container eventuell in die MAIL Api integrieren, nur Email Adressen die aus den Anwendungen kamen speichern und dann nach lookup Eingang zulassen , was nicht drin ist wird abgewiesen

##### Mail API , postfix commands als Go package abbilden ( reuseability, Lesbarkeit), Ansible Playbook fuer die API schreiben

###### Explizite Rest Routen um zwischen Rechnungen , Bestellungen etc zu differenzieren , entsprechend flaggen und Tag aus der Anwendung aus der sie kam in der DB setzen 

###### Mail Archiv abrufbar machen, Mail Archiv per cronjob Eintraege mit Aufbewahrungsdatum ueberschritten loeschen (soft delete), Fristen ermitteln je nach Typ 

##### Bounce Management implementieren , was passiert bei welchen Fehler Codes , Log ja/nein

#####  

###### Eventuell Logs speichern in der DB aber nicht alles , DATADOG Logging 

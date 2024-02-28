dot -Tpdf ./openEnrollment.dot  -o ./openEnrollment.pdf
dot -Tpdf ./Person_V1.dot  -o ./Person_V1.pdf
dot -Tpdf ./Person.dot  -o ./Person.pdf
dot -Gepsilon=.001 -Gstart=self -Tpdf ./BenefitsOverview.dot  -o ./BenefitsOverviewDot.pdf
neato -Gepsilon=.001 -Gstart=1000 -Tpdf ./BenefitsOverview.dot  -o ./BenefitsOverview.pdf
# neato -Tsvg ./BenefitsOverview.dot  -o ./BenefitsOverview.svg
docker compose  -f "./docker-compose-dev.yml" up
docker compose  -f -d "./docker-compose-working.yml" -d up
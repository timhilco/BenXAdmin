dot -Tpdf ./graphs/openEnrollment.dot  -o ./graphs/openEnrollment.pdf
dot -Gepsilon=.001 -Gstart=self -Tpdf ./BenefitsOverview.dot  -o ./BenefitsOverviewDot.pdf
neato -Gepsilon=.001 -Gstart=1000 -Tpdf ./BenefitsOverview.dot  -o ./BenefitsOverview.pdf
# neato -Tsvg ./BenefitsOverview.dot  -o ./BenefitsOverview.svg
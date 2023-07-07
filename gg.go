package main

type helmParam struct {
	name string
}

type helmResponse struct {
	message string
}

type helmDapr interface {
	Install(number int, param *helmParam) (bool, *helmResponse, error)
	UnInstall(number int) error
}

type HelmDapr struct {
	Conf string
}

func (m *HelmDapr) Install(number int, param *helmParam) (bool, *helmResponse, error) {
	m.Conf = "conf-HelmDapr"

	return true, &helmResponse{message: "installed success"}, nil
}

func (m *HelmDapr) UnInstall(number int) error {
	m.Conf = "conf-HelmDapr"

	return nil
}

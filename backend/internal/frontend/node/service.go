package frontendnode

type FrontendNodeService struct {
	FrontendNodeRepository *FrontendNodeRepository
}

func NewFrontendNodeService(frontendNodeRepository *FrontendNodeRepository) *FrontendNodeService {
	return &FrontendNodeService{
		FrontendNodeRepository: frontendNodeRepository,
	}
}

func (f *FrontendNodeService) GetComponents(componentType string) (*[]ComponentInfo, error) {
	return f.FrontendNodeRepository.GetComponents(componentType)
}

func (f *FrontendNodeService) GetComponentSchemaByName(componentName string) (any, error) {
	return f.FrontendNodeRepository.GetComponentSchemaByName(componentName)
}

package frontendnode

type FrontendNodeRepositoryInterface interface {
	GetComponents(componentType string) (*[]ComponentInfo, error)
	GetComponentSchemaByName(componentName string) (any, error)
	GetComponentUISchemaByName(componentName string) (any, error)
}

type FrontendNodeService struct {
	FrontendNodeRepository FrontendNodeRepositoryInterface
}

type FrontendNodeServiceInterface interface {
	GetComponents(componentType string) (*[]ComponentInfo, error)
	GetComponentSchemaByName(componentName string) (any, error)
	GetComponentUISchemaByName(componentName string) (any, error)
}

func NewFrontendNodeService(frontendNodeRepositoryInterface FrontendNodeRepositoryInterface) *FrontendNodeService {
	return &FrontendNodeService{
		FrontendNodeRepository: frontendNodeRepositoryInterface,
	}
}

func (f *FrontendNodeService) GetComponents(componentType string) (*[]ComponentInfo, error) {
	return f.FrontendNodeRepository.GetComponents(componentType)
}

func (f *FrontendNodeService) GetComponentSchemaByName(componentName string) (any, error) {
	return f.FrontendNodeRepository.GetComponentSchemaByName(componentName)
}

func (f *FrontendNodeService) GetComponentUISchemaByName(componentName string) (any, error) {
	return f.FrontendNodeRepository.GetComponentUISchemaByName(componentName)
}

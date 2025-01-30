package handlerBasic

import (
	"github.com/driscollco-cluster/operator-1password/internal/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests")
}

var _ = Describe("Basic Handler Test", func() {
	var (
		mockController *gomock.Controller
		mockRequest    *mocks.MockRequest
		mockLog        *mocks.MockLog
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockRequest = mocks.NewMockRequest(mockController)
		mockLog = mocks.NewMockLog(mockController)
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("Basic handler", func() {
		When("a request comes in", func() {
			When("the parameter 'one' is populated in the URL", func() {
				It("should produce a log and some output which starts with the URL argument", func() {
					mockRequest.EXPECT().Log().Return(mockLog)
					mockLog.EXPECT().Child("handler", "basic").Return(mockLog)
					mockRequest.EXPECT().ArgExists("one").Return(true)
					mockLog.EXPECT().Info("gcp project", "project", "")
					mockRequest.EXPECT().GetArg("one").Return("first argument").Times(2)
					mockLog.EXPECT().Info("got url argument", "one", "first argument")
					mockRequest.EXPECT().Success("received: first argument")

					Handle(mockRequest)
				})
			})
		})
	})
})

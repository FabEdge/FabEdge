package types_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/fabedge/fabedge/pkg/operator/types"
)

var _ = Describe("AgentArgumentMap", func() {
	It("can set, get and delete key and value pairs", func() {
		argMap := types.NewAgentArgumentMap()

		Expect(argMap.HasKey("hello")).To(BeFalse())
		Expect(argMap.Get("hello")).To(Equal(""))

		argMap.Set("hello", "world")

		Expect(argMap.HasKey("hello")).To(BeTrue())
		Expect(argMap.Get("hello")).To(Equal("world"))

		argMap.Delete("hello")
		Expect(argMap.HasKey("hello")).To(BeFalse())
		Expect(argMap.Get("hello")).To(Equal(""))
	})

	It("IsIPAMEnabled return true only if key 'enable-ipam' exists and has value 'true'", func() {
		argMap := types.NewAgentArgumentMap()

		key := "enable-ipam"
		argMap.Set(key, "")
		Expect(argMap.IsProxyEnabled()).To(BeFalse())

		argMap.Set(key, "true")
		Expect(argMap.IsIPAMEnabled()).To(BeTrue())
	})

	It("IsProxyEnabled return true only if 'enable-proxy' exists and has value 'true'", func() {
		argMap := types.NewAgentArgumentMap()

		key := "enable-proxy"
		argMap.Set(key, "")
		Expect(argMap.IsProxyEnabled()).To(BeFalse())

		argMap.Set(key, "true")
		Expect(argMap.IsProxyEnabled()).To(BeTrue())
	})

	It("ArgumentArray will generate an argument array sorted by argument name", func() {
		argMap := types.NewAgentArgumentMap()

		argMap.Set("enable-proxy", "false")
		argMap.Set("enable-ipam", "true")

		argumentArray := argMap.ArgumentArray()
		Expect(argumentArray).To(Equal([]string{
			"--enable-ipam=true",
			"--enable-proxy=false",
		}))
	})

	It("ArgumentArray always put log-level at the end of array and use 'v' replace it", func() {
		argMap := types.NewAgentArgumentMap()

		argMap.Set("log-level", "3")
		argMap.Set("enable-ipam", "true")

		argumentArray := argMap.ArgumentArray()
		Expect(argumentArray).To(Equal([]string{
			"--enable-ipam=true",
			"--v=3",
		}))
	})

	It("ArgumentArray will generate an argument array sorted by argument name", func() {
		argMap := types.NewAgentArgumentMap()

		argMap.Set("log-level", "3")
		argMap.Set("enable-ipam", "true")
		argMap.Set("enable-proxy", "false")

		argumentArray := argMap.ArgumentArray()
		Expect(argumentArray).To(Equal([]string{
			"--enable-ipam=true",
			"--enable-proxy=false",
			"--v=3",
		}))
	})
})

var _ = Describe("NewAgentArgumentMapFromEnv", func() {
	BeforeEach(func() {
		os.Setenv("AGENT_ARG_LOG_LEVEL", "3")
		os.Setenv("AGENT_ARG_ENABLE_IPAM", "true")
		os.Setenv("AGENT_ARG_ENABLE_PROXY", "")
	})

	AfterEach(func() {
		os.Unsetenv("AGENT_ARG_LOG_LEVEL")
		os.Unsetenv("AGENT_ARG_ENABLE_IPAM")
		os.Unsetenv("AGENT_ARG_ENABLE_PROXY")
	})

	It("build an AgentArgumentMap from environment variables which have prefix 'AGENT_ARG_'", func() {
		argMap := types.NewAgentArgumentMapFromEnv()

		Expect(len(argMap)).To(Equal(3))
	})

	It("each environment variable will be saved but its key will has prefix 'AGENT_ARG_' removed and lowered and all '_' are replaces with '-' ", func() {
		argMap := types.NewAgentArgumentMapFromEnv()

		Expect(argMap.Get("log-level")).To(Equal("3"))
		Expect(argMap.IsIPAMEnabled()).To(BeTrue())
		Expect(argMap.IsProxyEnabled()).To(BeFalse())
	})
})

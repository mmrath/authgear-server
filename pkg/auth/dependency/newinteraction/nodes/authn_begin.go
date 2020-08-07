package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationBegin{})
}

type EdgeAuthenticationBegin struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage: e.Stage,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage                newinteraction.AuthenticationStage `json:"stage"`
	Identity             *identity.Info                     `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig       `json:"-"`
	Authenticators       []*authenticator.Info              `json:"-"`
}

func (n *NodeAuthenticationBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.Identity = graph.MustGetUserLastIdentity()
	n.AuthenticationConfig = ctx.Config.Authentication
	n.Authenticators = ais
	return nil
}

func (n *NodeAuthenticationBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return n.deriveEdges(), nil
}

func (n *NodeAuthenticationBegin) deriveEdges() []newinteraction.Edge {
	var edges []newinteraction.Edge
	var availableAuthenticators []*authenticator.Info

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
			authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
		)
		newinteraction.SortAuthenticators(
			n.AuthenticationConfig.PrimaryAuthenticators,
			availableAuthenticators,
			func(i int) authn.AuthenticatorType {
				return availableAuthenticators[i].Type
			},
		)
	case newinteraction.AuthenticationStageSecondary:
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
		)
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	passwords := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypePassword),
	)
	totps := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeTOTP),
	)
	oobs := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeOOB),
	)

	if len(passwords) > 0 {
		edges = append(edges, &EdgeAuthenticationPassword{
			Stage:          n.Stage,
			Authenticators: passwords,
		})
	}

	if len(totps) > 0 {
		edges = append(edges, &EdgeAuthenticationTOTP{
			Stage:          n.Stage,
			Authenticators: totps,
		})
	}

	if len(oobs) > 0 {
		edges = append(edges, &EdgeAuthenticationOOBTrigger{
			Stage:          n.Stage,
			Authenticators: oobs,
		})
	}

	// No authenticators found, skip the authentication stage
	if len(edges) == 0 {
		edges = append(edges, &EdgeAuthenticationEnd{
			Stage:    n.Stage,
			Optional: true,
		})
	}

	// TODO(interaction): support choosing authenticator to use
	return edges[:1]
}

func (n *NodeAuthenticationBegin) AuthenticatorTypes() []authn.AuthenticatorType {
	edges := n.deriveEdges()

	var types []authn.AuthenticatorType
	for _, e := range edges {
		if e, ok := e.(interface {
			AuthenticatorType() authn.AuthenticatorType
		}); ok {
			types = append(types, e.AuthenticatorType())
		}
	}
	return types
}

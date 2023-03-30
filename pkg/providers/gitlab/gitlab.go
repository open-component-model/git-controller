package gitlab

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
	"github.com/open-component-model/git-controller/pkg/providers/gogit"
)

const (
	tokenKey      = "token"
	providerType  = "gitlab"
	defaultDomain = gitlab.DefaultDomain
)

var (
	// for now, only personal tokens are supported
	tokenType = "personal"
)

// Client gitlab.
type Client struct {
	client client.Client
	next   providers.Provider
}

// NewClient creates a new Gitlab client.
func NewClient(client client.Client, next providers.Provider) *Client {
	return &Client{
		client: client,
		next:   next,
	}
}

var _ providers.Provider = &Client{}

func (c *Client) CreateRepository(ctx context.Context, obj mpasv1alpha1.Repository) error {
	if obj.Spec.Provider != providerType {
		if c.next == nil {
			return fmt.Errorf("can't handle provider type '%s' and no next provider is configured", obj.Spec.Provider)
		}

		return c.next.CreateRepository(ctx, obj)
	}

	secret := &v1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      obj.Spec.Credentials.SecretRef.Name,
		Namespace: obj.Namespace,
	}, secret); err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	token, ok := secret.Data[tokenKey]
	if !ok {
		return fmt.Errorf("token '%s' not found in secret", tokenKey)
	}

	domain := defaultDomain
	if obj.Spec.Domain != "" {
		domain = obj.Spec.Domain
	}

	gc, err := gitlab.NewClient(string(token), tokenType, gitprovider.WithDomain(domain))
	if err != nil {
		return fmt.Errorf("failed to create github client: %w", err)
	}

	if obj.Spec.IsOrganization {
		return gogit.CreateOrganizationRepository(ctx, gc, domain, obj.Spec)
	}

	return gogit.CreateUserRepository(ctx, gc, domain, obj.Spec)
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, branch, description string) error {
	return nil
}

package github

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
)

const (
	tokenKey      = "token"
	providerType  = "github"
	defaultDomain = "github.com"
)

// Client github.
type Client struct {
	client client.Client
	next   providers.Provider
}

// TODO: Use this instead and somehow abstract the two clients.
type RepositoryOpts struct {
	Owner      string
	Domain     string
	Visibility gitprovider.RepositoryVisibility
}

// NewClient creates a new GitHub client.
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

	authenticationOption, err := c.constructAuthenticationOption(ctx, obj)
	if err != nil {
		return err
	}

	gc, err := github.NewClient(authenticationOption)
	if err != nil {
		return fmt.Errorf("failed to create github client: %w", err)
	}

	visibility := gitprovider.RepositoryVisibility(obj.Spec.Visibility)

	if err := gitprovider.ValidateRepositoryVisibility(visibility); err != nil {
		return fmt.Errorf("failed to validate visibility: %w", err)
	}

	domain := defaultDomain
	if obj.Spec.Domain != "" {
		domain = obj.Spec.Domain
	}

	if obj.Spec.IsOrganization {
		return c.createOrganizationRepository(ctx, gc, domain, visibility, obj.Spec)
	}

	return c.createUserRepository(ctx, gc, domain, visibility, obj.Spec)
}

// constructAuthenticationOption will take the object and construct an authentication option.
// For now, only token secret is supported, this will be extended in the future.
func (c *Client) constructAuthenticationOption(ctx context.Context, obj mpasv1alpha1.Repository) (gitprovider.ClientOption, error) {
	secret := &v1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      obj.Spec.Credentials.SecretRef.Name,
		Namespace: obj.Namespace,
	}, secret); err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	token, ok := secret.Data[tokenKey]
	if !ok {
		return nil, fmt.Errorf("token '%s' not found in secret", tokenKey)
	}

	return gitprovider.WithOAuth2Token(string(token)), nil
}

func (c *Client) createOrganizationRepository(ctx context.Context, gc gitprovider.Client, domain string, visibility gitprovider.RepositoryVisibility, spec mpasv1alpha1.RepositorySpec) error {
	logger := log.FromContext(ctx)

	ref := gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: spec.Owner,
		},
		RepositoryName: spec.RepositoryName,
	}
	info := gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar("main"),
		Visibility:    &visibility,
	}

	switch spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		if _, err := gc.OrgRepositories().Create(ctx, ref, info); err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		logger.Info("successfully created organization repository", "domain", domain, "repository", spec.RepositoryName)
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		_, created, err := gc.OrgRepositories().Reconcile(ctx, ref, info)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", spec.RepositoryName)
		} else {
			logger.Info("successfully created organization repository", "domain", domain, "repository", spec.RepositoryName)
		}
	default:
		return fmt.Errorf("unknown repository policy '%s'", spec.ExistingRepositoryPolicy)
	}

	return nil
}

func (c *Client) createUserRepository(ctx context.Context, gc gitprovider.Client, domain string, visibility gitprovider.RepositoryVisibility, spec mpasv1alpha1.RepositorySpec) error {
	logger := log.FromContext(ctx)

	ref := gitprovider.UserRepositoryRef{
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: spec.Owner,
		},
		RepositoryName: spec.RepositoryName,
	}
	info := gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar("main"),
		Visibility:    &visibility,
	}

	switch spec.ExistingRepositoryPolicy {
	case mpasv1alpha1.ExistingRepositoryPolicyFail:
		if _, err := gc.UserRepositories().Create(ctx, ref, info); err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}

		logger.Info("successfully created user repository", "domain", domain, "repository", spec.RepositoryName)
	case mpasv1alpha1.ExistingRepositoryPolicyAdopt:
		_, created, err := gc.UserRepositories().Reconcile(ctx, ref, info)
		if err != nil {
			return fmt.Errorf("failed to reconcile repository: %w", err)
		}

		if !created {
			logger.Info("using existing repository", "domain", domain, "repository", spec.RepositoryName)
		} else {
			logger.Info("successfully created user repository", "domain", domain, "repository", spec.RepositoryName)
		}
	default:
		return fmt.Errorf("unknown repository policy '%s'", spec.ExistingRepositoryPolicy)
	}

	return nil
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, branch, description string) error {
	return nil
}

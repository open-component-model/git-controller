<h1>OCM Controller API reference v1alpha1</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#delivery.ocm.software%2fv1alpha1">delivery.ocm.software/v1alpha1</a>
</li>
</ul>
<h2 id="delivery.ocm.software/v1alpha1">delivery.ocm.software/v1alpha1</h2>
<p>Package v1alpha1 contains API Schema definitions for the delivery v1alpha1 API group</p>
Resource Types:
<ul class="simple"></ul>
<h3 id="delivery.ocm.software/v1alpha1.CommitTemplate">CommitTemplate
</h3>
<p>
(<em>Appears on:</em>
<a href="#delivery.ocm.software/v1alpha1.SyncSpec">SyncSpec</a>)
</p>
<p>CommitTemplate defines the details of the commit to the external repository.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>email</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>message</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>targetBranch</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>baseBranch</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="delivery.ocm.software/v1alpha1.PullRequestTemplate">PullRequestTemplate
</h3>
<p>
(<em>Appears on:</em>
<a href="#delivery.ocm.software/v1alpha1.SyncSpec">SyncSpec</a>)
</p>
<p>PullRequestTemplate provides information for the created pull request.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>title</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>description</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>base</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="delivery.ocm.software/v1alpha1.Sync">Sync
</h3>
<p>Sync is the Schema for the syncs API</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#delivery.ocm.software/v1alpha1.SyncSpec">
SyncSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>snapshotRef</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repositoryRef</code><br>
<em>
<a href="https://pkg.go.dev/github.com/fluxcd/pkg/apis/meta#NamespacedObjectReference">
github.com/fluxcd/pkg/apis/meta.NamespacedObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>interval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>commitTemplate</code><br>
<em>
<a href="#delivery.ocm.software/v1alpha1.CommitTemplate">
CommitTemplate
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>subPath</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>prune</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>automaticPullRequestCreation</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>pullRequestTemplate</code><br>
<em>
<a href="#delivery.ocm.software/v1alpha1.PullRequestTemplate">
PullRequestTemplate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#delivery.ocm.software/v1alpha1.SyncStatus">
SyncStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="delivery.ocm.software/v1alpha1.SyncSpec">SyncSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#delivery.ocm.software/v1alpha1.Sync">Sync</a>)
</p>
<p>SyncSpec defines the desired state of Sync</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>snapshotRef</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>repositoryRef</code><br>
<em>
<a href="https://pkg.go.dev/github.com/fluxcd/pkg/apis/meta#NamespacedObjectReference">
github.com/fluxcd/pkg/apis/meta.NamespacedObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>interval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>commitTemplate</code><br>
<em>
<a href="#delivery.ocm.software/v1alpha1.CommitTemplate">
CommitTemplate
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>subPath</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>prune</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>automaticPullRequestCreation</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>pullRequestTemplate</code><br>
<em>
<a href="#delivery.ocm.software/v1alpha1.PullRequestTemplate">
PullRequestTemplate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="delivery.ocm.software/v1alpha1.SyncStatus">SyncStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#delivery.ocm.software/v1alpha1.Sync">Sync</a>)
</p>
<p>SyncStatus defines the observed state of Sync</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>digest</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>observedGeneration</code><br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>ObservedGeneration is the last reconciled generation.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Condition">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>pullRequestID</code><br>
<em>
int
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>

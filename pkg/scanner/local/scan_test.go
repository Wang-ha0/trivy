package local

import (
	"errors"
	"testing"

	ospkgDetector "github.com/aquasecurity/trivy/pkg/detector/ospkg"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	ftypes "github.com/aquasecurity/fanal/types"
	dtypes "github.com/aquasecurity/go-dep-parser/pkg/types"
	"github.com/aquasecurity/trivy/pkg/report"
	"github.com/aquasecurity/trivy/pkg/types"
)

func TestScanner_Scan(t *testing.T) {
	type args struct {
		target   string
		layerIDs []string
		options  types.ScanOptions
	}
	tests := []struct {
		name                    string
		args                    args
		applyLayersExpectation  ApplyLayersExpectation
		ospkgDetectExpectations []OspkgDetectorDetectExpectation
		libDetectExpectations   []LibraryDetectorDetectExpectation
		wantResults             report.Results
		wantOS                  *ftypes.OS
		wantEosl                bool
		wantErr                 string
	}{
		{
			name: "happy path",
			args: args{
				target:   "alpine:latest",
				layerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				options:  types.ScanOptions{VulnType: []string{"os", "library"}},
			},
			applyLayersExpectation: ApplyLayersExpectation{
				Args: ApplyLayersArgs{
					LayerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				},
				Returns: ApplyLayersReturns{
					Detail: ftypes.ImageDetail{
						OS: &ftypes.OS{
							Family: "alpine",
							Name:   "3.11",
						},
						Packages: []ftypes.Package{
							{Name: "musl", Version: "1.2.3"},
						},
						Applications: []ftypes.Application{
							{
								Type:     "bundler",
								FilePath: "/app/Gemfile.lock",
								Libraries: []dtypes.Library{
									{Name: "rails", Version: "6.0"},
								},
							},
						},
					},
				},
			},
			ospkgDetectExpectations: []OspkgDetectorDetectExpectation{
				{
					Args: OspkgDetectorDetectArgs{
						OsFamily: "alpine",
						OsName:   "3.11",
						Pkgs: []ftypes.Package{
							{Name: "musl", Version: "1.2.3"},
						},
					},
					Returns: OspkgDetectorDetectReturns{
						DetectedVulns: []types.DetectedVulnerability{
							{
								VulnerabilityID:  "CVE-2020-9999",
								PkgName:          "musl",
								InstalledVersion: "1.2.3",
								FixedVersion:     "1.2.4",
							},
						},
						Eosl: false,
					},
				},
			},
			libDetectExpectations: []LibraryDetectorDetectExpectation{
				{
					Args: LibraryDetectorDetectArgs{
						FilePath: "/app/Gemfile.lock",
						Pkgs: []dtypes.Library{
							{Name: "rails", Version: "6.0"},
						},
					},
					Returns: LibraryDetectorDetectReturns{
						DetectedVulns: []types.DetectedVulnerability{
							{
								VulnerabilityID:  "CVE-2020-10000",
								PkgName:          "rails",
								InstalledVersion: "6.0",
								FixedVersion:     "6.1",
							},
						},
					},
				},
			},
			wantResults: report.Results{
				{
					Target: "alpine:latest (alpine 3.11)",
					Vulnerabilities: []types.DetectedVulnerability{
						{
							VulnerabilityID:  "CVE-2020-9999",
							PkgName:          "musl",
							InstalledVersion: "1.2.3",
							FixedVersion:     "1.2.4",
						},
					},
				},
				{
					Target: "/app/Gemfile.lock",
					Vulnerabilities: []types.DetectedVulnerability{
						{
							VulnerabilityID:  "CVE-2020-10000",
							PkgName:          "rails",
							InstalledVersion: "6.0",
							FixedVersion:     "6.1",
						},
					},
				},
			},
			wantOS: &ftypes.OS{
				Family: "alpine",
				Name:   "3.11",
			},
		},
		{
			name: "happy path with empty os",
			args: args{
				target:   "alpine:latest",
				layerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				options:  types.ScanOptions{VulnType: []string{"os", "library"}},
			},
			applyLayersExpectation: ApplyLayersExpectation{
				Args: ApplyLayersArgs{
					LayerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				},
				Returns: ApplyLayersReturns{
					Detail: ftypes.ImageDetail{
						OS: &ftypes.OS{},
						Applications: []ftypes.Application{
							{
								Type:     "bundler",
								FilePath: "/app/Gemfile.lock",
								Libraries: []dtypes.Library{
									{Name: "rails", Version: "6.0"},
								},
							},
						},
					},
				},
			},
			libDetectExpectations: []LibraryDetectorDetectExpectation{
				{
					Args: LibraryDetectorDetectArgs{
						FilePath: "/app/Gemfile.lock",
						Pkgs: []dtypes.Library{
							{Name: "rails", Version: "6.0"},
						},
					},
					Returns: LibraryDetectorDetectReturns{
						DetectedVulns: []types.DetectedVulnerability{
							{
								VulnerabilityID:  "CVE-2020-10000",
								PkgName:          "rails",
								InstalledVersion: "6.0",
								FixedVersion:     "6.1",
							},
						},
					},
				},
			},
			wantResults: report.Results{
				{
					Target: "/app/Gemfile.lock",
					Vulnerabilities: []types.DetectedVulnerability{
						{
							VulnerabilityID:  "CVE-2020-10000",
							PkgName:          "rails",
							InstalledVersion: "6.0",
							FixedVersion:     "6.1",
						},
					},
				},
			},
			wantOS: &ftypes.OS{},
		},
		{
			name: "happy path with unknown os",
			args: args{
				target:   "alpine:latest",
				layerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				options:  types.ScanOptions{VulnType: []string{"os", "library"}},
			},
			applyLayersExpectation: ApplyLayersExpectation{
				Args: ApplyLayersArgs{
					LayerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				},
				Returns: ApplyLayersReturns{
					Detail: ftypes.ImageDetail{
						OS: &ftypes.OS{
							Family: "fedora",
							Name:   "27",
						},
						Applications: []ftypes.Application{
							{
								Type:     "bundler",
								FilePath: "/app/Gemfile.lock",
								Libraries: []dtypes.Library{
									{Name: "rails", Version: "6.0"},
								},
							},
						},
					},
				},
			},
			ospkgDetectExpectations: []OspkgDetectorDetectExpectation{
				{
					Args: OspkgDetectorDetectArgs{
						OsFamily: "fedora",
						OsName:   "27",
					},
					Returns: OspkgDetectorDetectReturns{
						Err: ospkgDetector.ErrUnsupportedOS,
					},
				},
			},
			libDetectExpectations: []LibraryDetectorDetectExpectation{
				{
					Args: LibraryDetectorDetectArgs{
						FilePath: "/app/Gemfile.lock",
						Pkgs: []dtypes.Library{
							{Name: "rails", Version: "6.0"},
						},
					},
					Returns: LibraryDetectorDetectReturns{
						DetectedVulns: []types.DetectedVulnerability{
							{
								VulnerabilityID:  "CVE-2020-10000",
								PkgName:          "rails",
								InstalledVersion: "6.0",
								FixedVersion:     "6.1",
							},
						},
					},
				},
			},
			wantResults: report.Results{
				{
					Target: "/app/Gemfile.lock",
					Vulnerabilities: []types.DetectedVulnerability{
						{
							VulnerabilityID:  "CVE-2020-10000",
							PkgName:          "rails",
							InstalledVersion: "6.0",
							FixedVersion:     "6.1",
						},
					},
				},
			},
			wantOS: &ftypes.OS{
				Family: "fedora",
				Name:   "27",
			},
		},
		{
			name: "happy path with only library detection",
			args: args{
				target:   "alpine:latest",
				layerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				options:  types.ScanOptions{VulnType: []string{"library"}},
			},
			applyLayersExpectation: ApplyLayersExpectation{
				Args: ApplyLayersArgs{
					LayerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				},
				Returns: ApplyLayersReturns{
					Detail: ftypes.ImageDetail{
						OS: &ftypes.OS{
							Family: "alpine",
							Name:   "3.11",
						},
						Packages: []ftypes.Package{
							{Name: "musl", Version: "1.2.3"},
						},
						Applications: []ftypes.Application{
							{
								Type:     "bundler",
								FilePath: "/app/Gemfile.lock",
								Libraries: []dtypes.Library{
									{Name: "rails", Version: "5.1"},
								},
							},
							{
								Type:     "composer",
								FilePath: "/app/composer-lock.json",
								Libraries: []dtypes.Library{
									{Name: "laravel", Version: "6.0.0"},
								},
							},
						},
					},
				},
			},
			libDetectExpectations: []LibraryDetectorDetectExpectation{
				{
					Args: LibraryDetectorDetectArgs{
						FilePath: "/app/Gemfile.lock",
						Pkgs: []dtypes.Library{
							{Name: "rails", Version: "5.1"},
						},
					},
					Returns: LibraryDetectorDetectReturns{
						DetectedVulns: []types.DetectedVulnerability{
							{
								VulnerabilityID:  "CVE-2020-11111",
								PkgName:          "rails",
								InstalledVersion: "5.1",
								FixedVersion:     "5.2",
							},
						},
					},
				},
				{
					Args: LibraryDetectorDetectArgs{
						FilePath: "/app/composer-lock.json",
						Pkgs: []dtypes.Library{
							{Name: "laravel", Version: "6.0.0"},
						},
					},
					Returns: LibraryDetectorDetectReturns{
						DetectedVulns: []types.DetectedVulnerability{
							{
								VulnerabilityID:  "CVE-2020-22222",
								PkgName:          "laravel",
								InstalledVersion: "6.0.0",
								FixedVersion:     "6.1.0",
							},
						},
					},
				},
			},
			wantResults: report.Results{
				{
					Target: "/app/Gemfile.lock",
					Vulnerabilities: []types.DetectedVulnerability{
						{
							VulnerabilityID:  "CVE-2020-11111",
							PkgName:          "rails",
							InstalledVersion: "5.1",
							FixedVersion:     "5.2",
						},
					},
				},
				{
					Target: "/app/composer-lock.json",
					Vulnerabilities: []types.DetectedVulnerability{
						{
							VulnerabilityID:  "CVE-2020-22222",
							PkgName:          "laravel",
							InstalledVersion: "6.0.0",
							FixedVersion:     "6.1.0",
						},
					},
				},
			},
			wantOS: &ftypes.OS{
				Family: "alpine",
				Name:   "3.11",
			},
		},
		{
			name: "sad path: ApplyLayers returns an error",
			args: args{
				target:   "alpine:latest",
				layerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				options:  types.ScanOptions{VulnType: []string{"os", "library"}},
			},
			applyLayersExpectation: ApplyLayersExpectation{
				Args: ApplyLayersArgs{
					LayerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				},
				Returns: ApplyLayersReturns{
					Err: errors.New("error"),
				},
			},
			wantErr: "failed to apply layers",
		},
		{
			name: "sad path: ospkgDetector.Detect returns an error",
			args: args{
				target:   "alpine:latest",
				layerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				options:  types.ScanOptions{VulnType: []string{"os", "library"}},
			},
			applyLayersExpectation: ApplyLayersExpectation{
				Args: ApplyLayersArgs{
					LayerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				},
				Returns: ApplyLayersReturns{
					Detail: ftypes.ImageDetail{
						OS: &ftypes.OS{
							Family: "alpine",
							Name:   "3.11",
						},
						Packages: []ftypes.Package{
							{Name: "musl", Version: "1.2.3"},
						},
					},
				},
			},
			ospkgDetectExpectations: []OspkgDetectorDetectExpectation{
				{
					Args: OspkgDetectorDetectArgs{
						OsFamily: "alpine",
						OsName:   "3.11",
						Pkgs: []ftypes.Package{
							{Name: "musl", Version: "1.2.3"},
						},
					},
					Returns: OspkgDetectorDetectReturns{
						Err: errors.New("error"),
					},
				},
			},
			wantErr: "failed to scan OS packages",
		},
		{
			name: "sad path: libDetector.Detect returns an error",
			args: args{
				target:   "alpine:latest",
				layerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				options:  types.ScanOptions{VulnType: []string{"library"}},
			},
			applyLayersExpectation: ApplyLayersExpectation{
				Args: ApplyLayersArgs{
					LayerIDs: []string{"sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
				},
				Returns: ApplyLayersReturns{
					Detail: ftypes.ImageDetail{
						OS: &ftypes.OS{
							Family: "alpine",
							Name:   "3.11",
						},
						Packages: []ftypes.Package{
							{Name: "musl", Version: "1.2.3"},
						},
						Applications: []ftypes.Application{
							{
								Type:     "bundler",
								FilePath: "/app/Gemfile.lock",
								Libraries: []dtypes.Library{
									{Name: "rails", Version: "6.0"},
								},
							},
						},
					},
				},
			},
			libDetectExpectations: []LibraryDetectorDetectExpectation{
				{
					Args: LibraryDetectorDetectArgs{
						FilePath: "/app/Gemfile.lock",
						Pkgs: []dtypes.Library{
							{Name: "rails", Version: "6.0"},
						},
					},
					Returns: LibraryDetectorDetectReturns{
						Err: errors.New("error"),
					},
				},
			},
			wantErr: "failed to scan application libraries",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applier := new(MockApplier)
			applier.ApplyApplyLayersExpectation(tt.applyLayersExpectation)

			ospkgDetector := new(MockOspkgDetector)
			ospkgDetector.ApplyDetectExpectations(tt.ospkgDetectExpectations)

			libDetector := new(MockLibraryDetector)
			libDetector.ApplyDetectExpectations(tt.libDetectExpectations)

			s := NewScanner(applier, ospkgDetector, libDetector)
			gotResults, gotOS, gotEosl, err := s.Scan(tt.args.target, "", tt.args.layerIDs, tt.args.options)
			if tt.wantErr != "" {
				require.NotNil(t, err, tt.name)
				require.Contains(t, err.Error(), tt.wantErr, tt.name)
				return
			} else {
				require.NoError(t, err, tt.name)
			}

			assert.Equal(t, tt.wantResults, gotResults)
			assert.Equal(t, tt.wantOS, gotOS)
			assert.Equal(t, tt.wantEosl, gotEosl)

			applier.AssertExpectations(t)
			ospkgDetector.AssertExpectations(t)
			libDetector.AssertExpectations(t)
		})
	}
}

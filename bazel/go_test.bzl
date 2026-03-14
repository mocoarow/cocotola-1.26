load("@rules_go//go:def.bzl", "go_test")

def smart_go_test(name, srcs, **kwargs):
    """
    Custom go_test macro that automatically creates separate targets based on build tags.

    This macro reads the source files and creates separate test targets for different sizes:
    - Files with '//go:build small' -> small test target
    - Files with '//go:build medium' -> medium test target
    - Files with '//go:build large' -> large test target
    - Other files -> regular test target
    """

    # Separate files by build tags (simplified approach using file naming conventions)
    regular_srcs = []
    small_srcs = []
    medium_srcs = []
    large_srcs = []
    helper_srcs = []  # Helper files that should be included in all test targets

    for src in srcs:
        # This is a simplified approach - in practice, we categorize by file naming patterns
        # since Bazel doesn't easily allow file content reading during analysis phase
        if "_s_test.go" in src:  # small test files
            small_srcs.append(src)
        elif "_m_test.go" in src:  # medium test files
            medium_srcs.append(src)
        elif "_l_test.go" in src:  # large test files
            large_srcs.append(src)
        elif ("helper_test.go" in src or "_helper_test.go" in src or
              "init_test.go" in src or "_init_test.go" in src or
              "export_test.go" in src or "_export_test.go" in src or
              "mocks_test.go" in src):  # helper files
            helper_srcs.append(src)
        else:
            regular_srcs.append(src)

    # Create regular test target (if any non-tagged files exist)
    if regular_srcs:
        go_test(
            name = name,
            srcs = regular_srcs,
            **kwargs
        )

    # Create small test target
    if small_srcs:
        go_test(
            name = name + "_s_test",
            size = "small",
            srcs = small_srcs + helper_srcs,
            gotags = ["small"],
            **kwargs
        )

    # Create medium test target
    if medium_srcs:
        go_test(
            name = name + "_m_test",
            size = "medium",
            srcs = medium_srcs + helper_srcs,
            gotags = ["medium"],
            **kwargs
        )

    # Create large test target
    if large_srcs:
        go_test(
            name = name + "_l_test",
            size = "large",
            srcs = large_srcs + helper_srcs,
            gotags = ["large"],
            **kwargs
        )
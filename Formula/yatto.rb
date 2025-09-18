class Yatto < Formula
  desc "Terminal-based to-do application built with Bubble Tea"
  homepage "https://github.com/handlebargh/yatto"
  url "https://github.com/handlebargh/yatto/archive/refs/tags/v0.15.0.tar.gz"
  sha256 "f26896ad3ed339ea51e3eb8d11a8b1c0ce094d28bc6fe4ef0fb9a50daacee5c5"
  license "MIT"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"yatto"
  end

  test do
    # Test version output contains version information
    output = shell_output("#{bin}/yatto -version")
    assert_match "Version:", output
    assert_match "yatto", output
    
    # Test help output
    help_output = shell_output("#{bin}/yatto -help 2>&1", 1)
    assert_match "Usage of", help_output
  end
end

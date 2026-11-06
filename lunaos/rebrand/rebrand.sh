#!/bin/bash
# Replayable Omarchy -> LunaOS rename of a git work tree.
#
#   rebrand.sh <tree>          rewrite <tree> in place (clean git work tree required)
#   rebrand.sh --check <tree>  list remaining occurrences, exit 1 if there are any
#
# Exit codes: 0 ok, 1 --check found occurrences, 2 precondition (usage, not a
# clean git work tree), 3 unknown casing of "omarchy" (tree left unchanged),
# 4 a renamed path would collide with an existing one or with another renamed
# path (tree left unchanged), 5 an internal step failed (perl, git or mktemp;
# the tree may be half rewritten, discard it with git checkout/reset).
#
# <tree>/lunaos/ and .git/ are never entered; binary files only have their path
# renamed. Exceptions live in exceptions.txt; the LICENSE and upstream fixture
# skips and the update repository rewrite are in the perl program below.
set -uo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
EXCEPTIONS="$HERE/exceptions.txt"

FAILED=5

die() {
  echo "rebrand: $*" >&2
  exit "$FAILED"
}

usage() {
  echo "usage: rebrand.sh [--check] <tree>" >&2
  exit 2
}

check_only=0
if [[ ${1:-} == "--check" ]]; then
  check_only=1
  shift
fi
if (( $# == 1 )); then
  tree=$1
else
  usage
fi

if [[ ! -d $tree ]] || [[ $(git -C "$tree" rev-parse --is-inside-work-tree 2>/dev/null) != "true" ]]; then
  echo "rebrand: $tree is not a git work tree" >&2
  exit 2
fi
if [[ -n $(git -C "$tree" rev-parse --show-prefix) ]]; then
  echo "rebrand: $tree is not the top of its git work tree" >&2
  exit 2
fi

work=$(mktemp -d "${TMPDIR:-/tmp}/rebrand.XXXXXX") || die "cannot create a temporary directory"
trap 'rm -rf "$work"' EXIT

cat >"$work/rebrand.pl" <<'PERL'
use strict;
use warnings;

my ($mode, $excfile, $tree, $listfile) = @ARGV;

# Data: explicit substitutions (not exceptions), applied before the exceptions are
# masked, so a broad exception such as omacom/omarchy keeps every other upstream link.
my @REPO_SUBSTITUTIONS = (
  ['https://github.com/omacom/omarchy.git',      'https://github.com/gregpavoine/LunaOS.git'],
  ['https://github.com/omacom/omarchy/releases', 'https://github.com/gregpavoine/LunaOS/releases'],
);
# Data: paths (relative to the tree) whose contents are never rewritten.
my %SKIPPED = ('LICENSE' => 1);
# Data: directories (relative to the tree) whose contents are never rewritten:
# verbatim copies of files shipped by upstream packages.
my @SKIPPED_DIRS = ('test/shell.d/fixtures/omarchy-mac-boot-mkinitcpio/');

my @exc;
open my $fh, '<', $excfile or die "cannot read $excfile: $!";
while (my $line = <$fh>) {
  chomp $line;
  next if $line =~ /^\s*(?:#|$)/;
  my ($pattern) = split /\t/, $line, 2;
  push @exc, $pattern if length $pattern;
}
close $fh;
@exc = sort { length($b) <=> length($a) || $a cmp $b } @exc;
my %id;
for my $i (0 .. $#exc) { $id{$exc[$i]} = $i unless exists $id{$exc[$i]}; }
my $alt = join '|', map { quotemeta } @exc;
my $exc_re = @exc ? qr/(?:$alt)/ : qr/(?!)/;

sub protect {
  my ($s) = @_;
  $s =~ s/($exc_re)/'@@LUNAOS_KEEP_' . $id{$1} . '@@'/ge;
  return $s;
}

sub restore {
  my ($s) = @_;
  $s =~ s/\@\@LUNAOS_KEEP_(\d+)\@\@/$exc[$1]/g;
  return $s;
}

sub repo_substitute {
  my ($s) = @_;
  for my $pair (@REPO_SUBSTITUTIONS) {
    my ($from, $to) = @$pair;
    $s =~ s/\Q$from\E/$to/g;
  }
  return $s;
}

sub casings {
  my ($s) = @_;
  $s =~ s/omarchy/lunaos/g;
  $s =~ s/Omarchy/LunaOS/g;
  $s =~ s/OMARCHY/LUNAOS/g;
  return $s;
}

sub unknown_casing {
  my ($s) = @_;
  for my $m ($s =~ /(omarchy)/ig) {
    return 1 unless $m eq 'omarchy' || $m eq 'Omarchy' || $m eq 'OMARCHY';
  }
  return 0;
}

sub skipped {
  my ($p) = @_;
  return 1 if $p =~ m{^lunaos/} || $SKIPPED{$p};
  for my $d (@SKIPPED_DIRS) { return 1 if index($p, $d) == 0; }
  return 0;
}

sub read_list {
  open my $l, '<:raw', $listfile or die "cannot read $listfile: $!";
  local $/ = "\0";
  my @items = <$l>;
  close $l;
  chomp @items;
  return @items;
}

sub slurp {
  my ($path) = @_;
  open my $in, '<:raw', $path or return undef;
  local $/;
  my $s = <$in>;
  close $in;
  return $s;
}

if ($mode eq 'contents-unknown' || $mode eq 'contents-check') {
  for my $f (read_list()) {
    next if skipped($f) || -l "$tree/$f";
    my $s = slurp("$tree/$f");
    next unless defined $s;
    my @lines = split /\n/, ($mode eq 'contents-unknown' ? protect(repo_substitute($s)) : protect($s)), -1;
    for my $i (0 .. $#lines) {
      if ($mode eq 'contents-unknown') {
        print "$f:" . ($i + 1) . "\n" if unknown_casing($lines[$i]);
      } elsif ($lines[$i] =~ /omarchy/i) {
        print "$f:" . ($i + 1) . ":" . restore($lines[$i]) . "\n";
      }
    }
  }
} elsif ($mode eq 'rewrite') {
  my $count = 0;
  for my $f (read_list()) {
    next if skipped($f) || -l "$tree/$f";
    my $s = slurp("$tree/$f");
    next unless defined $s;
    my $t = restore(casings(protect(repo_substitute($s))));
    next if $t eq $s;
    open my $out, '>:raw', "$tree/$f" or die "cannot write $f: $!";
    print $out $t;
    close $out;
    $count++;
  }
  print "$count\n";
} elsif ($mode eq 'paths-unknown') {
  for my $p (read_list()) {
    next if skipped($p);
    print "$p\n" if unknown_casing(protect($p));
  }
} elsif ($mode eq 'paths-check') {
  for my $p (read_list()) {
    next if skipped($p);
    print "$p:path name\n" if protect($p) =~ /omarchy/i;
  }
} elsif ($mode eq 'paths-plan') {
  # One line per path (file or directory) whose last component changes,
  # deepest first, so parents are still under their old name when renamed.
  my (%seen, @plan);
  for my $p (read_list()) {
    next if skipped($p);
    my @c = split m{/}, $p;
    for my $k (0 .. $#c) {
      my $prefix = join '/', @c[0 .. $k];
      next if $seen{$prefix}++;
      next unless $c[$k] =~ /omarchy/i;
      my $renamed = restore(casings(protect($prefix)));
      my ($last) = $renamed =~ m{([^/]*)$};
      next if $last eq $c[$k];
      my $new = $k > 0 ? join('/', @c[0 .. $k - 1], $last) : $last;
      push @plan, [$k, $prefix, $new];
    }
  }
  for my $e (sort { $b->[0] <=> $a->[0] || $a->[1] cmp $b->[1] } @plan) {
    print "$e->[1]\t$e->[2]\n";
  }
} else {
  die "unknown mode $mode";
}
PERL

# run_perl <mode> <list>: output goes to stdout; a perl failure is fatal.
run_perl() {
  perl "$work/rebrand.pl" "$1" "$EXCEPTIONS" "$tree" "$2" || die "internal failure in step $1"
}

# Text files that mention omarchy (binary files are excluded by -I), and all tracked paths.
# git grep exits 1 when nothing matches; anything above 1 is an error.
grep_rc=0
git -C "$tree" grep -z -lI -i omarchy -- . ':!lunaos' >"$work/contents.lst" || grep_rc=$?
(( grep_rc <= 1 )) || die "git grep failed ($grep_rc)"
git -C "$tree" ls-files -z >"$work/paths.lst" || die "git ls-files failed"

if (( check_only )); then
  run_perl contents-check "$work/contents.lst" >"$work/remaining.txt"
  run_perl paths-check "$work/paths.lst" >>"$work/remaining.txt"
  if [[ -s $work/remaining.txt ]]; then
    cat "$work/remaining.txt"
    exit 1
  else
    echo "rebrand --check: no occurrence left"
    exit 0
  fi
fi

status=$(git -C "$tree" status --porcelain) || die "git status failed"
if [[ -n $status ]]; then
  echo "rebrand: $tree has uncommitted changes" >&2
  exit 2
fi

# Decide everything before touching the tree.
run_perl contents-unknown "$work/contents.lst" >"$work/unknown.txt"
run_perl paths-unknown "$work/paths.lst" >>"$work/unknown.txt"
if [[ -s $work/unknown.txt ]]; then
  echo "rebrand: unknown casing of omarchy (expected omarchy, Omarchy or OMARCHY):" >&2
  cat "$work/unknown.txt" >&2
  exit 3
fi

run_perl paths-plan "$work/paths.lst" >"$work/plan.txt"
collision=0
while IFS=$'\t' read -r old new; do
  if [[ -e $tree/$new ]]; then
    echo "rebrand: cannot rename $old to $new: target exists" >&2
    collision=1
  fi
done <"$work/plan.txt"
cut -f2 "$work/plan.txt" | sort | uniq -d >"$work/duplicates.txt" || die "cannot check duplicate targets"
while IFS= read -r new; do
  echo "rebrand: several paths would be renamed to $new" >&2
  collision=1
done <"$work/duplicates.txt"
if (( collision )); then
  exit 4
fi

modified=$(run_perl rewrite "$work/contents.lst") || die "rewrite failed"

renamed=0
while IFS=$'\t' read -r old new; do
  git -C "$tree" mv -- "$old" "$new" || die "git mv $old $new failed"
  renamed=$((renamed + 1))
done <"$work/plan.txt"

echo "rebrand: renamed $renamed paths, modified $modified files"

/bin/sh
set -eux

# Auto-detect the matching module from the /opt/wallarm set by running the embedded script.
module_path="$(cd /opt/wallarm && /opt/wallarm/pick-module.sh)"
if [ -z "${module_path}" ] ; then
    echo "No matching module found in Wallarm installer"
    exit 1
fi

# The module name (recipe id) is the last directory.
module_name="${module_path%/*}"
module_name="${module_name##*/}"

# Take the matching module.
mkdir -p /usr/lib/nginx/modules
mv "/opt/wallarm/modules/${module_name}/"* -t "/usr/lib/nginx/modules/"

# Take also the libwallarm if split-static module was built (WALLARM_STATIC=2).
if ls /opt/wallarm/modules/libwallarm.so* 2> /dev/null ; then
    mkdir -p /usr/local/lib
    mv /opt/wallarm/modules/libwallarm.so* -t "/usr/local/lib/"
fi

# Free up disk space.
rm -rf /opt/wallarm/modules

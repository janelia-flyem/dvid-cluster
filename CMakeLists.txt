CMAKE_MINIMUM_REQUIRED(VERSION 2.8)
project (dvid-cluster)

include (ExternalProject)

set (RUN_ENVIRONMENT "Workstation" CACHE TYPE STRING)
if (NOT CMAKE_BUILD_TYPE)
    set (CMAKE_BUILD_TYPE Release)
endif ()

# This should only be built on the cluster where we have access to the grid engine.
# Check for expected environment variables.  Will have to be modified if not running
# on SGE cluster.
if (NOT DEFINED ENV{SGE_CLUSTER_NAME})
    message(FATAL_ERROR "Can only build cluster programs on cluster-accessible hosts.  Aborting.")
endif ()

# Require a GOPATH
if (NOT DEFINED ENV{GOPATH})
    message(FATAL_ERROR "GOPATH is required to be set before building.  Aborting.")
endif ()

################################################################################
# Check if BUILDEM_DIR has already been assigned.  If not, create a default.
set (BUILDEM_DIR "None" CACHE TYPE STRING)

if (${BUILDEM_DIR} STREQUAL "None")
    message (FATAL_ERROR "ERROR: Buildem directory (for all downloads & builds) should be specified via -DBUILDEM_DIR=<path> on cmake command line.")
endif ()

message ("BUILDEM downloads and builds will be placed here: ${BUILDEM_DIR}")

###############################################################################

###############################################################################
# Download and install buildem, if it isn't already in BUILDEM_DIR.
set (BUILDEM_REPO_DIR ${BUILDEM_DIR}/src/buildem)
if (NOT EXISTS ${BUILDEM_REPO_DIR}/python.cmake)
    message ("Installing buildem repo...")
    ExternalProject_Add(buildem
        PREFIX ${BUILDEM_DIR}
        GIT_REPOSITORY https://github.com/janelia-flyem/buildem.git
        UPDATE_COMMAND ""
        PATCH_COMMAND ""
        CONFIGURE_COMMAND "" 
        BUILD_COMMAND ""
        BUILD_IN_SOURCE 1
        INSTALL_COMMAND ""
    )
    message ("\n**********************************************************\n")
    message ("\nAfter running make, you must re-run the cmake command once")
    message ("buildem has been downloaded!\n")
    message ("\n***********************************************************\n")
else ()
    ###############################################################################

    # Use modules from the downloaded buildem
    set (CMAKE_MODULE_PATH ${BUILDEM_REPO_DIR})
    message("Using cmake modules from ${BUILDEM_REPO_DIR}")

    # Download and compile dependencies.
    # NOTE: To prevent build of Go, you can set the environment variable CI_ENV.  This is
    #  particularly useful when using a continuous integration system with its own reasonable
    #  Go toolset.
    if (DEFINED ENV{CI_ENV})
        message("Building under $ENV{CI_ENV} continuous integration system.  Using standard Go.")
        set (golang_NAME "")
    else ()
        message("Including Go language build...")
        include (golang)
    endif ()

    include (snappy)


    set (CGO_FLAGS
    CGO_CFLAGS="-I${BUILDEM_INCLUDE_DIR};-I$ENV{SGE_ROOT}/include";CGO_LDFLAGS="-L${BUILDEM_LIB_DIR};-L$ENV{SGE_ROOT}/lib/lx-amd64;-lsnappy;-ldrmaa")

    add_custom_target (go-drmaa-install
        ${BUILDEM_ENV_STRING} ${CGO_FLAGS} go get ${GO_GET} github.com/janelia-flyem/drmaa
        DEPENDS     ${golang_NAME}
        COMMENT     "Retrieved Go DRMAA job submission library")

    add_custom_target (go-drmaa
        git submodule init
        COMMAND     ./build.sh
        DEPENDS     go-drmaa-install
        WORKING_DIRECTORY $ENV{GOPATH}/src/github.com/janelia-flyem/drmaa
        COMMENT     "Built Go DRMAA job submission library")
    
    add_custom_target (dvid
        ${BUILDEM_ENV_STRING} ${CGO_FLAGS} go get ${GO_GET} github.com/janelia-flyem/dvid
        DEPENDS     ${golang_NAME}
        COMMENT     "Added dvid package")

    add_custom_target (gopackages-install
        ${BUILDEM_ENV_STRING} ${CGO_FLAGS} go get ${GO_GET} github.com/janelia-flyem/go
        DEPENDS     ${golang_NAME} ${snappy_NAME}
        COMMENT     "Added go package dependencies for janelia-flyem projects")

    add_custom_target (gopackages
        git submodule init
        COMMAND     git submodule update
        DEPENDS     gopackages-install
        WORKING_DIRECTORY $ENV{GOPATH}/src/github.com/janelia-flyem/go
        COMMENT     "Updated required Go packages")

    add_custom_target (groupcache
        ${BUILDEM_ENV_STRING} ${CGO_FLAGS} go get ${GO_GET} github.com/golang/groupcache
        DEPENDS     ${golang_NAME}
        COMMENT     "Added groupcache dependency")

    # Make manager + node
    add_custom_target (all
        echo "Building manager and node executable"
        DEPENDS     manager node)
    
    # Build manager
    add_custom_target (manager
        ${BUILDEM_ENV_STRING} ${GO_ENV} ${CGO_FLAGS} go build -o ${BUILDEM_BIN_DIR}/dvid-manager 
            -v main.go 
        WORKING_DIRECTORY   ${CMAKE_CURRENT_SOURCE_DIR}/manager-exe
        DEPENDS     ${golang_NAME} go-drmaa gopackages 
        COMMENT     "Compiled and installed DVID cluster manager.")

    # Build node program
    add_custom_target (node
        ${BUILDEM_ENV_STRING} ${GO_ENV} ${CGO_FLAGS} go build -o ${BUILDEM_BIN_DIR}/dvid-node 
            -v main.go 
        WORKING_DIRECTORY   ${CMAKE_CURRENT_SOURCE_DIR}/node-exe
        DEPENDS     ${golang_NAME} gopackages groupcache
        COMMENT     "Compiled and installed DVID node program.")


   # Add testing
   add_custom_target (test
        ${BUILDEM_ENV_STRING} ${CGO_FLAGS} go test ${CMAKE_CURRENT_SOURCE_DIR}/node
        COMMENT    "Ran tests on DVID and node")


    ###############################################################################
endif()

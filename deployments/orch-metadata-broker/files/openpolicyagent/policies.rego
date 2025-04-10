# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package metadatav1

import future.keywords.in

hasWriteAccess {
    admRole := sprintf("%s_ao-rw", [input.metadata.activeprojectid[0]])
    ecmRole := sprintf("%s_cl-rw", [input.metadata.activeprojectid[0]])
    eimRole := sprintf("%s_im-rw", [input.metadata.activeprojectid[0]])

    some role in input.metadata["realm_access/roles"] # iteration
    [admRole, ecmRole, eimRole][_] == role
}

hasReadAccess {
    ecmRoleRead := sprintf("%s_cl-r", [input.metadata.activeprojectid[0]])
    eimRoleRead := sprintf("%s_im-r", [input.metadata.activeprojectid[0]])
    admRoleReadWrite := sprintf("%s_ao-rw", [input.metadata.activeprojectid[0]])
    ecmRoleReadWrite := sprintf("%s_cl-rw", [input.metadata.activeprojectid[0]])
    eimRoleReadWrite := sprintf("%s_im-rw", [input.metadata.activeprojectid[0]])

    some role in input.metadata["realm_access/roles"] # iteration
    [ecmRoleRead, eimRoleRead, admRoleReadWrite, ecmRoleReadWrite, eimRoleReadWrite][_] == role
}

CreateOrUpdateRequest {
    hasWriteAccess
}

DeleteRequest {
    hasWriteAccess
}

GetRequest {
    hasReadAccess
}

DeleteProjectRequest {
    hasWriteAccess
}
// Code generated by "stringer -type=SubStatus -linecomment"; DO NOT EDIT.

package dsupdate

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[UserIDNotSpecified-480]
	_ = x[PasswordNotSpecified-481]
	_ = x[MissingAParameter-482]
	_ = x[DomainNameNotSpecified-483]
	_ = x[InvalidDomainName-484]
	_ = x[InvalidUserID-485]
	_ = x[InvalidDigestAndDigestTypeCombination-486]
	_ = x[TheContentsOfAtLeastOneParameterIsSyntacticallyWrong-487]
	_ = x[AtLeastOneDSKeyHasAnInvalidAlgorithm-488]
	_ = x[InvalidSequenceOfSets-489]
	_ = x[UnknownParameterGiven-495]
	_ = x[UnknownUserID-496]
	_ = x[UnknownDomainName-497]
	_ = x[AuthenticationFailed-531]
	_ = x[AuthorizationFailed-532]
	_ = x[AuthenticatingUsingThisPasswordTypeIsNotSupported-533]
}

const (
	_SubStatus_name_0 = "user ID not specifiedpassword not specifiedmissing a parameterdomain name not specifiedinvalid domain nameinvalid user IDinvalid digest and digest type combinationthe contents of at least one parameter is syntactically wrongat least one DS key has an invalid algorithminvalid sequence of sets"
	_SubStatus_name_1 = "unknown parameter givenunknown user IDunknown domain name"
	_SubStatus_name_2 = "authentication failedauthorization failedauthenticating using this password type is not supported"
)

var (
	_SubStatus_index_0 = [...]uint16{0, 21, 43, 62, 87, 106, 121, 163, 224, 268, 292}
	_SubStatus_index_1 = [...]uint8{0, 23, 38, 57}
	_SubStatus_index_2 = [...]uint8{0, 21, 41, 97}
)

func (i SubStatus) String() string {
	switch {
	case 480 <= i && i <= 489:
		i -= 480
		return _SubStatus_name_0[_SubStatus_index_0[i]:_SubStatus_index_0[i+1]]
	case 495 <= i && i <= 497:
		i -= 495
		return _SubStatus_name_1[_SubStatus_index_1[i]:_SubStatus_index_1[i+1]]
	case 531 <= i && i <= 533:
		i -= 531
		return _SubStatus_name_2[_SubStatus_index_2[i]:_SubStatus_index_2[i+1]]
	default:
		return "SubStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
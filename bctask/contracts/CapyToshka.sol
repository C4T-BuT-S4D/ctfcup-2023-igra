// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";

contract CapyToshka is ERC721 {
    mapping(uint256 tokenId => string) private _secrets;
    uint256 _tokenId;

    constructor() ERC721("CapyToshka", "CAPY") {
    }

    function secret(uint256 tokenId) external view returns (string memory) {
        return _secrets[tokenId];
    }

    function mint(address to, string calldata secret_) external {
        uint256 tokenId = _tokenId;
        _tokenId++;
        _secrets[tokenId] = secret_;
        _safeMint(to, tokenId, "");
    }
}

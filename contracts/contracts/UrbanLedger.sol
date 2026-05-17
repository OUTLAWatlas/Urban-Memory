// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title UrbanMemory Ledger
 * @dev Creates an immutable chain of custody for municipal GIS data using SHA-256 hashes.
 */
contract UrbanLedger {
    
    // State variables
    address public admin;

    // Struct to hold the cryptographic proof of a map layer
    struct LayerCommitment {
        string sha256Hash;   // The cryptographic fingerprint of the GeoJSON
        string ipfsCID;      // Decentralized backup pointer on IPFS
        uint256 timestamp;   // When it was anchored to the blockchain
        string sourceRef;    // e.g., "BMC Development Plan 1991"
        bool exists;         // To check if a record exists
    }

    // Mapping: Layer Type (e.g., "zone_industrial") => Year (e.g., 1991) => Commitment Data
    mapping(string => mapping(uint16 => LayerCommitment)) public layerRegistry;

    // Events allow your Go backend to listen for when a new hash is locked in
    event HashCommitted(string layerType, uint16 year, string sha256Hash, string ipfsCID, uint256 timestamp);

    // Only the system admin (your backend) should be able to write to the ledger
    modifier onlyAdmin() {
        require(msg.sender == admin, "Unauthorized: Only the admin can commit hashes.");
        _;
    }

    constructor() {
        admin = msg.sender;
    }

    /**
     * @dev Commits a new spatial data hash to the blockchain.
     * @param _layerType The type of infrastructure (e.g., "forest_cover")
     * @param _year The historical year of the data (e.g., 1991)
     * @param _sha256Hash The calculated hash of the GeoJSON file
     * @param _sourceRef The authority the data came from
     */
    function commitLayerHash(
        string memory _layerType, 
        uint16 _year, 
        string memory _sha256Hash,
        string memory _ipfsCID,
        string memory _sourceRef
    ) public onlyAdmin {
        // Prevent overwriting history! Once a year's hash is locked, it cannot be changed.
        require(!layerRegistry[_layerType][_year].exists, "Tamper Alert: Hash for this layer and year already exists!");
        require(bytes(_ipfsCID).length > 0, "IPFS CID is required");

        layerRegistry[_layerType][_year] = LayerCommitment({
            sha256Hash: _sha256Hash,
            ipfsCID: _ipfsCID,
            timestamp: block.timestamp,
            sourceRef: _sourceRef,
            exists: true
        });

        emit HashCommitted(_layerType, _year, _sha256Hash, _ipfsCID, block.timestamp);
    }

    /**
     * @dev Fetches the trusted hash for a specific layer and year to verify against the database.
     */
    function verifyLayer(string memory _layerType, uint16 _year) public view returns (string memory, uint256, string memory, string memory) {
        require(layerRegistry[_layerType][_year].exists, "No cryptographic record found for this layer/year.");
        
        LayerCommitment memory record = layerRegistry[_layerType][_year];
        return (record.sha256Hash, record.timestamp, record.sourceRef, record.ipfsCID);
    }
}

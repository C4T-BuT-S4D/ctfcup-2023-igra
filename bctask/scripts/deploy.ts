import { formatEther, parseEther } from "viem";
import hre from "hardhat";

async function main() {
  const capy = await hre.viem.deployContract("CapyToshka");

  console.log(
    capy.address
  );
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});

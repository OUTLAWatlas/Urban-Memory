"use client"

import React, { useEffect, useState } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { GisNavigationPin } from "./gis-navigation-pin"

const FACTS: string[] = [
  "Jane Jacobs: Cities have the capability of providing something for everybody, only because, and only when, they are created by everybody.",
  "Urban facts: Over 55% of the world population now lives in cities — spatial data drives better services.",
  "GIS note: High-resolution spatial data increases planning accuracy but also storage and compute needs.",
  "Security: Cryptographic hashing ensures ledger entries remain tamper-evident across distributed systems.",
  "Design: Roads shape behaviour — small changes to the network change travel patterns dramatically.",
  "Metadata: Good metadata makes datasets reusable across departments and time horizons.",
  "Policy: Participatory mapping empowers communities to claim and protect public space."
]

export default function GlobalLoader() {
  const [index, setIndex] = useState(0)

  useEffect(() => {
    const id = setInterval(() => {
      setIndex((prev) => {
        if (FACTS.length <= 1) return 0
        let next = prev
        while (next === prev) {
          next = Math.floor(Math.random() * FACTS.length)
        }
        return next
      })
    }, 3500)
    return () => clearInterval(id)
  }, [])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-zinc-950 text-white">
      <div className="absolute inset-0 pointer-events-none opacity-50" aria-hidden>
        <div className="w-full h-full bg-gradient-to-b from-transparent via-zinc-900/40 to-black" />
      </div>

      <div className="absolute top-4 left-4 text-xs font-mono text-zinc-400">ESTABLISHING LEDGER CONNECTION...</div>

      <div className="relative flex flex-col items-center gap-8 px-6 py-12">
        <div className="absolute -z-10 w-[520px] h-[520px] rounded-full bg-gradient-to-r from-cyan-800/5 via-emerald-800/6 to-transparent blur-3xl" />

        <div className="relative flex flex-col items-center">
          <div className="relative flex items-end justify-center">
            <div className="w-56 h-56 sm:w-72 sm:h-72 rounded-full bg-zinc-900/60 backdrop-blur-sm border border-zinc-800/40 shadow-2xl overflow-hidden flex items-center justify-center">
              <div
                className="absolute inset-0"
                style={{
                  backgroundImage:
                    `radial-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
                     linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
                     linear-gradient(90deg, rgba(255,255,255,0.01) 1px, transparent 1px)`,
                  backgroundSize: "24px 24px, 24px 24px, 100% 100%",
                  backgroundPosition: "0 0, 12px 12px, 0 0"
                }}
              />

              <motion.div
                className="absolute inset-0 rounded-full pointer-events-none"
                aria-hidden
                style={{
                  mixBlendMode: "screen",
                  opacity: 0.9,
                }}
                animate={{ rotate: 360 }}
                transition={{ repeat: Infinity, duration: 6, ease: "linear" }}
              >
                <div
                  className="absolute inset-0 rounded-full"
                  style={{
                    background: `conic-gradient(from 0deg, rgba(34,211,238,0.14) 0deg, rgba(34,211,238,0.08) 25deg, transparent 45deg)`,
                    filter: "blur(6px)",
                    opacity: 0.95
                  }}
                />
                <div
                  className="absolute inset-0 rounded-full"
                  style={{
                    background: `conic-gradient(from 0deg, rgba(34,211,238,0.6) 0deg, rgba(34,211,238,0.18) 8deg, rgba(34,211,238,0.02) 18deg, transparent 40deg)`,
                    mixBlendMode: "screen",
                    opacity: 0.8
                  }}
                />
              </motion.div>

              <motion.div
                className="absolute inset-0 rounded-full pointer-events-none"
                style={{ mixBlendMode: "screen" }}
                animate={{ rotate: 360 }}
                transition={{ repeat: Infinity, duration: 6, ease: "linear" }}
              >
                <div
                  className="absolute inset-0 rounded-full"
                  style={{
                    background: "radial-gradient(ellipse at 30% 40%, rgba(34,211,238,0.12) 0%, transparent 20%)",
                    filter: "blur(18px)",
                    opacity: 0.9
                  }}
                />
              </motion.div>

              <div className="w-6 h-6 rounded-full bg-cyan-400/8 border border-cyan-400/10" />
            </div>

            <motion.div
              className="absolute -top-10 flex items-end justify-center"
              animate={{ rotateY: [0, 360] }}
              transition={{ repeat: Infinity, duration: 3.5, ease: "linear" }}
            >
              <motion.div
                animate={{ y: [0, -10, 0] }}
                transition={{ repeat: Infinity, duration: 2.4, ease: "easeInOut" }}
                className="pointer-events-none"
              >
                <div className="w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center">
                  <GisNavigationPin />
                </div>
              </motion.div>
            </motion.div>
          </div>
        </div>

        <div className="mt-6 w-full max-w-2xl text-center">
          <div className="mx-auto max-w-xl px-4">
            <AnimatePresence mode="wait">
              <motion.p
                key={index}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -6 }}
                transition={{ duration: 0.5 }}
                className="text-sm text-zinc-300/95 leading-relaxed"
              >
                {FACTS[index]}
              </motion.p>
            </AnimatePresence>
          </div>
        </div>
      </div>
    </div>
  )
}
